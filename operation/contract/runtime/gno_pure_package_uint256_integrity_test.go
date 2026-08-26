package runtime

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const onblocUint256PackagePath = "gno.land/p/onbloc/uint256"

type gnoSourceManifest struct {
	Schema        string                  `json:"schema"`
	PackagePath   string                  `json:"package_path"`
	HashAlgorithm string                  `json:"hash_algorithm"`
	BundleFraming string                  `json:"bundle_framing"`
	Files         []gnoSourceManifestFile `json:"files"`
	BundleSHA256  string                  `json:"bundle_sha256"`
}

type gnoSourceManifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func TestOnblocUint256SourceManifestIntegrity(t *testing.T) {
	packageDir := onblocUint256SourceDir(t)
	manifest := readOnblocUint256Manifest(t, packageDir)

	if manifest.PackagePath != onblocUint256PackagePath {
		t.Fatalf("unexpected package identity %q", manifest.PackagePath)
	}
	if manifest.Schema != "smart-contract-model.gno-source-bundle/v1" {
		t.Fatalf("unexpected source manifest schema %q", manifest.Schema)
	}
	if manifest.HashAlgorithm != "sha256" {
		t.Fatalf("unexpected source hash algorithm %q", manifest.HashAlgorithm)
	}

	wantPaths := make([]string, 0, len(manifest.Files))
	contents := make(map[string][]byte, len(manifest.Files))
	for i, file := range manifest.Files {
		if filepath.IsAbs(file.Path) || filepath.ToSlash(file.Path) != file.Path || file.Path != filepath.Base(file.Path) {
			t.Fatalf("manifest file path must be a package-root-relative slash path: %q", file.Path)
		}
		if i > 0 && strings.Compare(manifest.Files[i-1].Path, file.Path) >= 0 {
			t.Fatalf("manifest files are not in byte-wise lexicographic order")
		}
		body, err := os.ReadFile(filepath.Join(packageDir, filepath.FromSlash(file.Path)))
		if err != nil {
			t.Fatalf("read vendored production file %q: %v", file.Path, err)
		}
		sum := sha256.Sum256(body)
		if got := hex.EncodeToString(sum[:]); got != file.SHA256 {
			t.Fatalf("SHA-256 mismatch for %q: got %s, want %s", file.Path, got, file.SHA256)
		}
		wantPaths = append(wantPaths, file.Path)
		contents[file.Path] = body
	}

	gotPaths := productionGnoFiles(t, packageDir)
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("vendored production file set differs from manifest: got %v, want %v", gotPaths, wantPaths)
	}
	if got := canonicalGnoSourceBundleHash(wantPaths, contents); got != manifest.BundleSHA256 {
		t.Fatalf("canonical bundle SHA-256 mismatch: got %s, want %s", got, manifest.BundleSHA256)
	}
}

func TestOnblocUint256ProductionBundleExcludesNonProductionFiles(t *testing.T) {
	packageDir := onblocUint256SourceDir(t)
	for _, name := range []string{
		"arithmetic_test.gno",
		"bitwise_test.gno",
		"cmp_test.gno",
		"conversion_test.gno",
		"uint256_test.gno",
		"gnomod.toml",
	} {
		if _, err := os.Stat(filepath.Join(packageDir, name)); !os.IsNotExist(err) {
			t.Fatalf("excluded upstream file %q is present in production source directory", name)
		}
	}
}

func TestOnblocUint256MetadataDoesNotPinHostFilesystemPaths(t *testing.T) {
	packageDir := onblocUint256SourceDir(t)
	for _, name := range []string{"SOURCE_MANIFEST.json", "SOURCE_PROVENANCE.md"} {
		body, err := os.ReadFile(filepath.Join(packageDir, name))
		if err != nil {
			t.Fatalf("read source metadata %q: %v", name, err)
		}
		for _, forbidden := range []string{"/Users/", "GOMODCACHE", "GNOROOT", "pkg/mod/"} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("source metadata %q exposes host path marker %q", name, forbidden)
			}
		}
	}
}

func onblocUint256SourceDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate uint256 integrity test source")
	}
	return filepath.Join(filepath.Dir(filename), "gno_pure_packages", filepath.FromSlash(onblocUint256PackagePath))
}

func readOnblocUint256Manifest(t *testing.T, packageDir string) gnoSourceManifest {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(packageDir, "SOURCE_MANIFEST.json"))
	if err != nil {
		t.Fatalf("read uint256 source manifest: %v", err)
	}
	var manifest gnoSourceManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("decode uint256 source manifest: %v", err)
	}
	return manifest
}

func productionGnoFiles(t *testing.T, packageDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		t.Fatalf("read vendored uint256 package: %v", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gno") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)
	return files
}

func canonicalGnoSourceBundleHash(paths []string, contents map[string][]byte) string {
	h := sha256.New()
	var length [8]byte
	for _, name := range paths {
		binary.BigEndian.PutUint64(length[:], uint64(len([]byte(name))))
		_, _ = h.Write(length[:])
		_, _ = h.Write([]byte(name))
		body := contents[name]
		binary.BigEndian.PutUint64(length[:], uint64(len(body)))
		_, _ = h.Write(length[:])
		_, _ = h.Write(body)
	}
	return hex.EncodeToString(h.Sum(nil))
}
