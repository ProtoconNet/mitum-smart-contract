package runtime

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type uint256ConsensusBundle struct {
	Schema         string `json:"schema"`
	PublicIdentity struct {
		PackagePath string `json:"package_path"`
		PackageName string `json:"package_name"`
		PackageType string `json:"package_type"`
	} `json:"public_identity"`
	DependencyIdentity struct {
		Module  string `json:"module"`
		Version string `json:"version"`
	} `json:"dependency_identity"`
	SemanticIdentity struct {
		NativeSemantics        string `json:"native_semantics"`
		GasSchedule            string `json:"gas_schedule"`
		CanonicalDecimalPolicy string `json:"canonical_decimal_policy"`
		SchemaRuleset          string `json:"schema_ruleset"`
		SnapshotCodec          string `json:"snapshot_codec"`
		SnapshotVersion        uint64 `json:"snapshot_version"`
	} `json:"semantic_identity"`
	HashIdentity struct {
		WrapperSHA256      string `json:"wrapper_sha256"`
		OnblocBundleSHA256 string `json:"onbloc_bundle_sha256"`
	} `json:"hash_identity"`
	ResourceIdentity struct {
		WriteGas, QueryGas, WriteAllocationBytes, QueryAllocationBytes, CanonicalToHexGas, MuldivGas, SqrtGas, CallContractGas, TransferSenderToContractGas, TransferContractToGas int64
		DepthLimit, TouchedContractLimit                                                                                                                                           int
	} `json:"-"`
	ResourceRaw                 map[string]int64  `json:"resource_identity"`
	SourceFiles                 map[string]string `json:"source_files"`
	CanonicalSourceBundleSHA256 string            `json:"canonical_source_bundle_sha256"`
}

func readUint256ConsensusBundle(t *testing.T) uint256ConsensusBundle {
	t.Helper()
	b, err := os.ReadFile("UINT256_CONSENSUS_BUNDLE.json")
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err = json.Unmarshal(b, &top); err != nil {
		t.Fatal(err)
	}
	if err = rejectDuplicateSourcePaths(top["source_files"]); err != nil {
		t.Fatal(err)
	}
	var m uint256ConsensusBundle
	if err = json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func rejectDuplicateSourcePaths(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return fmt.Errorf("source_files must be an object")
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		path, ok := token.(string)
		if !ok {
			return fmt.Errorf("source_files key is not a string")
		}
		if _, found := seen[path]; found {
			return fmt.Errorf("duplicate source path %q", path)
		}
		seen[path] = struct{}{}
		var hash string
		if err := decoder.Decode(&hash); err != nil {
			return err
		}
	}
	return nil
}

var expectedUint256ConsensusSourcePaths = []string{
	"go.mod",
	"go.sum",
	"operation/contract/runtime/engine.go",
	"operation/contract/runtime/execution_session.go",
	"operation/contract/runtime/gno_engine.go",
	"operation/contract/runtime/gno_execution_context.go",
	"operation/contract/runtime/gno_hostabi_package.go",
	"operation/contract/runtime/gno_limits.go",
	"operation/contract/runtime/gno_native_gas.go",
	"operation/contract/runtime/gno_native_resolver.go",
	"operation/contract/runtime/gno_pure_package.go",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/arithmetic.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/bits_table.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/bitwise.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/cmp.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/conversion.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/error.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/mod.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/uint256.gno",
	"operation/contract/runtime/gno_pure_packages/gno.land/p/onbloc/uint256/utils.gno",
	"operation/contract/runtime/gno_pure_packages/mitum/math/v1/u256/u256.gno",
	"operation/contract/runtime/gno_snapshot_codec.go",
	"operation/contract/runtime/gno_uint256_native.go",
	"operation/contract/runtime/import_policy.go",
	"operation/contract/runtime/schema_ruleset.go",
	"operation/contract/runtime/state_readers.go",
}

func consensusRepoFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

func TestUint256ConsensusBundleIdentity(t *testing.T) {
	m := readUint256ConsensusBundle(t)
	if m.Schema != "smart-contract-model.uint256-consensus-bundle/v1" {
		t.Fatal(m.Schema)
	}
	if m.PublicIdentity.PackagePath != Uint256PackagePath || m.PublicIdentity.PackageName != "u256" || m.PublicIdentity.PackageType != "MPStdlibProd" {
		t.Fatalf("public identity=%+v", m.PublicIdentity)
	}
	pkg, err := readEmbeddedGnoPurePackage(Uint256PackagePath)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Type != gno.MPStdlibProd {
		t.Fatal(pkg.Type)
	}
	if m.DependencyIdentity.Module != "github.com/gnolang/gno" || !strings.Contains(string(consensusRepoFile(t, "go.mod")), m.DependencyIdentity.Module+" "+m.DependencyIdentity.Version) {
		t.Fatalf("dependency=%+v", m.DependencyIdentity)
	}
	if m.HashIdentity.WrapperSHA256 != Uint256WrapperSourceSHA256 || m.HashIdentity.OnblocBundleSHA256 != "2e60329f269f003bc64c023aeb4f3288ca597e2cf906819b37f46a6de0d7aeb5" {
		t.Fatalf("hash identity=%+v", m.HashIdentity)
	}
	if IsAllowedTypedContractImport(OnblocUint256PackagePath) || IsAllowedTypedContractImport("mitum/math/u256/v1") || !IsAllowedTypedContractImport(Uint256PackagePath) {
		t.Fatal("import identity policy mismatch")
	}
}

func TestUint256ConsensusBundleSourceHashes(t *testing.T) {
	m := readUint256ConsensusBundle(t)
	if len(m.SourceFiles) != len(expectedUint256ConsensusSourcePaths) {
		t.Fatalf("source count=%d, want %d", len(m.SourceFiles), len(expectedUint256ConsensusSourcePaths))
	}
	expected := make(map[string]struct{}, len(expectedUint256ConsensusSourcePaths))
	for _, path := range expectedUint256ConsensusSourcePaths {
		if _, found := expected[path]; found {
			t.Fatalf("duplicate expected source path %q", path)
		}
		expected[path] = struct{}{}
		if _, found := m.SourceFiles[path]; !found {
			t.Fatalf("manifest missing expected source path %q", path)
		}
	}
	for path := range m.SourceFiles {
		if _, found := expected[path]; !found {
			t.Fatalf("manifest contains unexpected source path %q", path)
		}
	}
	for p, want := range m.SourceFiles {
		sum := sha256.Sum256(consensusRepoFile(t, p))
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Fatalf("%s=%s want %s", p, got, want)
		}
	}
}

func consensusBundleHash(t *testing.T, files map[string]string) string {
	t.Helper()
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Slice(paths, func(i, j int) bool { return strings.Compare(paths[i], paths[j]) < 0 })
	h := sha256.New()
	var n [8]byte
	for _, p := range paths {
		b := consensusRepoFile(t, p)
		binary.BigEndian.PutUint64(n[:], uint64(len([]byte(p))))
		h.Write(n[:])
		h.Write([]byte(p))
		binary.BigEndian.PutUint64(n[:], uint64(len(b)))
		h.Write(n[:])
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}
func TestUint256ConsensusBundleHashDeterministic(t *testing.T) {
	m := readUint256ConsensusBundle(t)
	a, b := consensusBundleHash(t, m.SourceFiles), consensusBundleHash(t, m.SourceFiles)
	if a != b || a != m.CanonicalSourceBundleSHA256 {
		t.Fatalf("bundle %s %s want %s", a, b, m.CanonicalSourceBundleSHA256)
	}
}

func TestUint256ConsensusBundleRuntimeConstants(t *testing.T) {
	m := readUint256ConsensusBundle(t)
	want := map[string]int64{"write_gas": GnoWriteGasLimit, "query_gas": GnoQueryGasLimit, "write_allocation_bytes": GnoWriteMaxAllocBytes, "query_allocation_bytes": GnoQueryMaxAllocBytes, "canonical_to_hex_gas": uint256NativeCanonicalToHexGasBase, "muldiv_gas": uint256NativeMulDivGasBase, "sqrt_gas": uint256NativeSqrtGasBase, "call_contract_gas": mitumNativeCallContractGasBase, "transfer_sender_to_contract_gas": mitumNativeCurrencyTransferGasBase, "transfer_contract_to_gas": mitumNativeCurrencyTransferGasBase, "depth_limit": MaxContractCallDepth, "touched_contract_limit": MaxTouchedContractsPerOperation}
	if !reflect.DeepEqual(m.ResourceRaw, want) {
		t.Fatalf("resources=%v want %v", m.ResourceRaw, want)
	}
	s := m.SemanticIdentity
	if s.NativeSemantics != Uint256NativeSemanticsVersion || s.GasSchedule != Uint256NativeGasSchedule || s.CanonicalDecimalPolicy != Uint256CanonicalDecimalPolicy || s.SchemaRuleset != CurrentSchemaRulesetVersion || s.SnapshotCodec != GnoSnapshotCodecName || s.SnapshotVersion != GnoSnapshotVersion {
		t.Fatalf("semantic identity=%+v", s)
	}
}
