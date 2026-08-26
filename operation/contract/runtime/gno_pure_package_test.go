package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
)

var onblocUint256ProductionFiles = []string{
	"arithmetic.gno",
	"bits_table.gno",
	"bitwise.gno",
	"cmp.gno",
	"conversion.gno",
	"error.gno",
	"mod.gno",
	"uint256.gno",
	"utils.gno",
}

func TestEmbeddedOnblocUint256MemPackage(t *testing.T) {
	pkg, err := readEmbeddedGnoPurePackage(OnblocUint256PackagePath)
	if err != nil {
		t.Fatalf("readEmbeddedGnoPurePackage returned error: %v", err)
	}
	if pkg.Name != "uint256" || pkg.Path != OnblocUint256PackagePath || pkg.Type != gno.MPUserProd {
		t.Fatalf("unexpected pure package identity: name=%q path=%q type=%v", pkg.Name, pkg.Path, pkg.Type)
	}
	if got := memPackageFileNames(pkg); !reflect.DeepEqual(got, onblocUint256ProductionFiles) {
		t.Fatalf("unexpected production files: %v", got)
	}
	for _, file := range pkg.Files {
		for _, forbidden := range []string{"_test.gno", "_filetest.gno", "gnomod.toml", "LICENSE", "SOURCE_"} {
			if strings.Contains(file.Name, forbidden) {
				t.Fatalf("MemPackage includes metadata or test file %q", file.Name)
			}
		}
	}
}

func TestEmbeddedUint256WrapperIdentity(t *testing.T) {
	pkg, err := readEmbeddedGnoPurePackage(Uint256PackagePath)
	if err != nil {
		t.Fatalf("read wrapper: %v", err)
	}
	if pkg.Name != "u256" || pkg.Path != "mitum/math/v1/u256" || pkg.Type != gno.MPStdlibProd {
		t.Fatalf("unexpected wrapper identity: name=%q path=%q type=%v", pkg.Name, pkg.Path, pkg.Type)
	}
	if got := memPackageFileNames(pkg); !reflect.DeepEqual(got, []string{"u256.gno"}) {
		t.Fatalf("unexpected wrapper files: %v", got)
	}
	if err := gno.ValidateMemPackageAny(pkg); err != nil {
		t.Fatalf("validate wrapper identity: %v", err)
	}
	sum := sha256.Sum256([]byte(pkg.Files[0].Body))
	if got := hex.EncodeToString(sum[:]); got != Uint256WrapperSourceSHA256 {
		t.Fatalf("wrapper source SHA-256 = %s, want %s", got, Uint256WrapperSourceSHA256)
	}
}

func TestGnoPureMemPackagesAreDeterministicAndCloneIsolated(t *testing.T) {
	source := "package contract\nimport \"" + Uint256PackagePath + "\"\n"
	first, err := GnoPureMemPackagesForContract(source)
	if err != nil {
		t.Fatalf("GnoPureMemPackagesForContract returned error: %v", err)
	}
	second, err := GnoPureMemPackagesForContract(source)
	if err != nil {
		t.Fatalf("second GnoPureMemPackagesForContract returned error: %v", err)
	}
	if !reflect.DeepEqual(stdlibPackageShape(first), stdlibPackageShape(second)) {
		t.Fatalf("pure runtime package order is not deterministic")
	}
	last := first[len(first)-1]
	if last.Path != Uint256PackagePath {
		t.Fatalf("expected pure root last, got %q", last.Path)
	}
	first[len(first)-1].Files[0].Body = "mutated"
	third, err := GnoPureMemPackagesForContract(source)
	if err != nil {
		t.Fatalf("third GnoPureMemPackagesForContract returned error: %v", err)
	}
	if third[len(third)-1].Files[0].Body == "mutated" {
		t.Fatal("caller mutated canonical pure package cache")
	}
	if packages, err := GnoPureMemPackagesForContract("package contract\n"); err != nil || len(packages) != 0 {
		t.Fatalf("contract without pure import loaded packages: %#v, %v", packages, err)
	}
}

func TestGnoPureRuntimeResolverSelectsOnlyExactCanonicalPath(t *testing.T) {
	for _, importPath := range []string{
		OnblocUint256PackagePath,
		"mitum/math/u256/v1",
		"mitum/math/v1",
		"mitum/math/v1/u256/subpackage",
		"mitum/math/v1/u2562",
		"gno.land/p/onbloc",
		"gno.land/p/onbloc/uint256/subpackage",
		"gno.land/p/onbloc/uint2562",
		"gno.land/p/other/package",
		"gno.land/r/demo/realm",
	} {
		packages, err := GnoPureMemPackagesForContract("package contract\nimport _ \"" + importPath + "\"\n")
		if err != nil {
			t.Fatalf("runtime resolver failed while ignoring disallowed path %q: %v", importPath, err)
		}
		if len(packages) != 0 {
			t.Fatalf("runtime resolver selected packages for non-canonical path %q", importPath)
		}
	}
}

func TestOnblocUint256DependenciesLoadFirstWithoutDuplicates(t *testing.T) {
	packages, err := GnoPureMemPackagesForContract(
		"package contract\nimport \"" + Uint256PackagePath + "\"\n",
	)
	if err != nil {
		t.Fatalf("GnoPureMemPackagesForContract returned error: %v", err)
	}
	indices := map[string]int{}
	for i, pkg := range packages {
		if _, found := indices[pkg.Path]; found {
			t.Fatalf("duplicate runtime package %q", pkg.Path)
		}
		indices[pkg.Path] = i
	}
	rootIndex := indices[Uint256PackagePath]
	if indices[OnblocUint256PackagePath] >= rootIndex {
		t.Fatal("internal Onbloc dependency must load before public wrapper")
	}
	for _, dependency := range []string{"encoding/binary", "errors", "math/bits", "strconv", "strings"} {
		index, found := indices[dependency]
		if !found {
			t.Fatalf("missing uint256 stdlib dependency %q", dependency)
		}
		if index >= rootIndex {
			t.Fatalf("dependency %q loaded after pure root", dependency)
		}
	}
	for _, pkg := range packages[:rootIndex] {
		dependencies, err := gnoStdlibImportPaths(pkg)
		if err != nil {
			t.Fatalf("inspect stdlib dependency %q: %v", pkg.Path, err)
		}
		for _, dependency := range dependencies {
			dependencyIndex, found := indices[dependency]
			if !found {
				t.Fatalf("missing indirect stdlib dependency %q of %q", dependency, pkg.Path)
			}
			if dependencyIndex >= indices[pkg.Path] {
				t.Fatalf("indirect stdlib dependency %q loaded after %q", dependency, pkg.Path)
			}
		}
	}
	if IsAllowedTypedContractImport("math/bits") {
		t.Fatal("pure dependency unexpectedly widened direct contract stdlib policy")
	}
}

func TestGnoPurePackageGraphDeterministicDependencyFirst(t *testing.T) {
	packages := map[string]*gnostd.MemPackage{
		"p/a": testPurePackage("p/a", `package a; import "p/c"`),
		"p/b": testPurePackage("p/b", `package b; import "p/c"`),
		"p/c": testPurePackage("p/c", "package c"),
	}
	read := func(path string) (*gnostd.MemPackage, error) {
		return cloneGnoMemPackages([]*gnostd.MemPackage{packages[path]})[0], nil
	}
	limits := GnoPurePackageLimits{MaxPackageCount: 4, MaxDependencyDepth: 4, MaxSourceBytes: 1024, MaxFilesPerPackage: 2}
	allowed := stringSet([]string{"p/a", "p/b", "p/c"})
	first, _, err := resolveGnoPurePackageGraph([]string{"p/b", "p/a"}, allowed, read, limits)
	if err != nil {
		t.Fatalf("resolveGnoPurePackageGraph returned error: %v", err)
	}
	second, _, err := resolveGnoPurePackageGraph([]string{"p/a", "p/b"}, allowed, read, limits)
	if err != nil {
		t.Fatalf("second resolveGnoPurePackageGraph returned error: %v", err)
	}
	want := []string{"p/c", "p/a", "p/b"}
	if got := memPackagePaths(first); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected dependency-first order: %v", got)
	}
	if !reflect.DeepEqual(memPackagePaths(first), memPackagePaths(second)) {
		t.Fatal("graph order depends on root input order")
	}
}

func TestGnoPurePackageGraphRejectsCycleAndLimits(t *testing.T) {
	base := GnoPurePackageLimits{MaxPackageCount: 4, MaxDependencyDepth: 4, MaxSourceBytes: 1024, MaxFilesPerPackage: 2}
	tests := []struct {
		name     string
		roots    []string
		packages map[string]*gnostd.MemPackage
		limits   GnoPurePackageLimits
		want     string
	}{
		{name: "cycle", roots: []string{"p/a"}, packages: map[string]*gnostd.MemPackage{"p/a": testPurePackage("p/a", `package a; import "p/a"`)}, limits: base, want: "dependency cycle"},
		{name: "depth", roots: []string{"p/a"}, packages: map[string]*gnostd.MemPackage{"p/a": testPurePackage("p/a", `package a; import "p/b"`), "p/b": testPurePackage("p/b", "package b")}, limits: GnoPurePackageLimits{MaxPackageCount: 4, MaxDependencyDepth: 1, MaxSourceBytes: 1024, MaxFilesPerPackage: 2}, want: "dependency depth"},
		{name: "count", roots: []string{"p/a"}, packages: map[string]*gnostd.MemPackage{"p/a": testPurePackage("p/a", `package a; import "p/b"`), "p/b": testPurePackage("p/b", "package b")}, limits: GnoPurePackageLimits{MaxPackageCount: 1, MaxDependencyDepth: 4, MaxSourceBytes: 1024, MaxFilesPerPackage: 2}, want: "count exceeds"},
		{name: "bytes", roots: []string{"p/a"}, packages: map[string]*gnostd.MemPackage{"p/a": testPurePackage("p/a", "package a")}, limits: GnoPurePackageLimits{MaxPackageCount: 4, MaxDependencyDepth: 4, MaxSourceBytes: 1, MaxFilesPerPackage: 2}, want: "source bytes exceed"},
		{name: "files", roots: []string{"p/a"}, packages: map[string]*gnostd.MemPackage{"p/a": {Name: "a", Path: "p/a", Type: gno.MPUserProd, Files: []*gnostd.MemFile{{Name: "a.gno", Body: "package a"}, {Name: "b.gno", Body: "package a"}}}}, limits: GnoPurePackageLimits{MaxPackageCount: 4, MaxDependencyDepth: 4, MaxSourceBytes: 1024, MaxFilesPerPackage: 1}, want: "production files"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			read := func(path string) (*gnostd.MemPackage, error) { return tt.packages[path], nil }
			_, _, err := resolveGnoPurePackageGraph(tt.roots, stringSet(mapKeys(tt.packages)), read, tt.limits)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestEmbeddedPurePackageErrorsDoNotExposeHostPaths(t *testing.T) {
	_, err := readEmbeddedGnoPurePackage("gno.land/p/missing")
	if err == nil {
		t.Fatal("expected missing pure package error")
	}
	for _, forbidden := range []string{"/Users/", "GOMODCACHE", "GNOROOT", "pkg/mod/", "github.com/gnolang/gno@"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("pure package error exposes host path marker %q: %v", forbidden, err)
		}
	}
}

func testPurePackage(path string, source string) *gnostd.MemPackage {
	return &gnostd.MemPackage{Name: path[strings.LastIndex(path, "/")+1:], Path: path, Type: gno.MPUserProd, Files: []*gnostd.MemFile{{Name: "package.gno", Body: source}}}
}

func memPackageFileNames(pkg *gnostd.MemPackage) []string {
	out := make([]string, 0, len(pkg.Files))
	for _, file := range pkg.Files {
		out = append(out, file.Name)
	}
	return out
}

func memPackagePaths(packages []*gnostd.MemPackage) []string {
	out := make([]string, 0, len(packages))
	for _, pkg := range packages {
		out = append(out, pkg.Path)
	}
	return out
}

func mapKeys(values map[string]*gnostd.MemPackage) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	return out
}
