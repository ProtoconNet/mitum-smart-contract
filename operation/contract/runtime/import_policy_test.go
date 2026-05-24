package runtime

import (
	"reflect"
	"testing"
)

func TestAllowedTypedContractImportSpecsMatchCurrentPolicy(t *testing.T) {
	expected := []AllowedImportSpec{
		{Path: MitumChainPackagePath, Kind: AllowedImportHostABI},
		{Path: "strconv", Kind: AllowedImportStdlib},
		{Path: "strings", Kind: AllowedImportStdlib},
		{Path: "errors", Kind: AllowedImportStdlib},
		{Path: "bytes", Kind: AllowedImportStdlib},
		{Path: "encoding/hex", Kind: AllowedImportStdlib},
		{Path: "encoding/base64", Kind: AllowedImportStdlib},
		{Path: "unicode/utf8", Kind: AllowedImportStdlib},
	}

	if got := AllowedTypedContractImportSpecs(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected canonical import specs: %#v", got)
	}
}

func TestAllowedTypedContractImportSpecsDriveDerivedConsumers(t *testing.T) {
	allowedPaths := allowedTypedContractImportPaths()
	if got := CurrentSchemaRuleset().ImportRules.AllowedImports; !reflect.DeepEqual(got, allowedPaths) {
		t.Fatalf("schema ruleset imports %#v do not match canonical paths %#v", got, allowedPaths)
	}
	if !reflect.DeepEqual(AllowedTypedContractImports, allowedPaths) {
		t.Fatalf("compatibility allowed imports %#v do not match canonical paths %#v", AllowedTypedContractImports, allowedPaths)
	}

	expectedHostPaths := allowedTypedContractImportPathsByKind(AllowedImportHostABI)
	hostPackages := HostABIMemPackages()
	hostPaths := make([]string, 0, len(hostPackages))
	for _, pkg := range hostPackages {
		hostPaths = append(hostPaths, pkg.Path)
	}
	if !reflect.DeepEqual(hostPaths, expectedHostPaths) {
		t.Fatalf("host ABI packages %#v do not match canonical host paths %#v", hostPaths, expectedHostPaths)
	}

	expectedStdlibPaths := allowedTypedContractImportPathsByKind(AllowedImportStdlib)
	if len(expectedStdlibPaths) == 0 {
		t.Fatal("expected canonical stdlib import paths")
	}
	source := "package contract\nimport (\n"
	for _, importPath := range expectedStdlibPaths {
		source += "\t\"" + importPath + "\"\n"
	}
	source += ")\n"
	gotStdlibPaths, err := contractGnoStdlibImports(source)
	if err != nil {
		t.Fatalf("contractGnoStdlibImports returned error: %v", err)
	}
	if !reflect.DeepEqual(gotStdlibPaths, expectedStdlibPaths) {
		t.Fatalf("runtime stdlib roots %#v do not match canonical stdlib paths %#v", gotStdlibPaths, expectedStdlibPaths)
	}
}

func TestAllowedTypedContractImportSpecsAreDefensiveAndDisjoint(t *testing.T) {
	specs := AllowedTypedContractImportSpecs()
	seen := map[string]AllowedImportKind{}
	for _, spec := range specs {
		if prior, found := seen[spec.Path]; found {
			t.Fatalf("canonical import %q has duplicate kinds %q and %q", spec.Path, prior, spec.Kind)
		}
		seen[spec.Path] = spec.Kind
	}
	if got := seen[MitumChainPackagePath]; got != AllowedImportHostABI {
		t.Fatalf("expected %q to remain host ABI, got %q", MitumChainPackagePath, got)
	}

	specs[0].Path = "fmt"
	specs[0].Kind = AllowedImportStdlib
	fresh := AllowedTypedContractImportSpecs()
	if fresh[0].Path != MitumChainPackagePath || fresh[0].Kind != AllowedImportHostABI {
		t.Fatalf("canonical import specs were modified through returned slice: %#v", fresh)
	}
}
