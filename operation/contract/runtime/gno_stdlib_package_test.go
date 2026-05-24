package runtime

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
)

const allowedStdlibRuntimeContractSource = `package contract
import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"mitum/chain"
	"strconv"
	"strings"
	"unicode/utf8"
)

var result string

func Initialize(ctx chain.WriteContext) error {
	buf := bytes.NewBufferString("go")
	result = strings.ToUpper(buf.String())
	result = result + ":" + strconv.FormatInt(ctx.GetHeight(), 10)
	result = result + ":" + hex.EncodeToString([]byte{1, 2})
	result = result + ":" + base64.StdEncoding.EncodeToString([]byte("gno"))
	if !utf8.ValidString(result) {
		return errors.New("invalid utf8")
	}
	return nil
}

func StoreUpper(ctx chain.WriteContext, text string) error {
	result = strings.ToUpper(text)
	return nil
}

func GetResult(ctx chain.QueryContext) string {
	return result
}

func Encode(ctx chain.QueryContext, text string) string {
	return hex.EncodeToString([]byte(text)) + ":" + base64.StdEncoding.EncodeToString([]byte(text))
}

func ErrorText(ctx chain.QueryContext) string {
	return errors.New("stdlib-error").Error()
}
`

func TestAllowedGnoStdlibImportsExecuteInRuntime(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractstdlib01")
	sender := base.NewStringAddress("senderstdlib001")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(42),
		ContractCode: allowedStdlibRuntimeContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(42), result.StateMerges)

	assertStdlibQueryResult(t, engine, states, contract, sender, "GetResult", nil, "GO:42:0102:Z25v")
	assertStdlibQueryResult(t, engine, states, contract, sender, "Encode", map[string]string{"text": "hi"}, "6869:aGk=")
	assertStdlibQueryResult(t, engine, states, contract, sender, "ErrorText", nil, "stdlib-error")

	result, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(43),
		ContractCode: allowedStdlibRuntimeContractSource,
		Function:     "StoreUpper",
		CallData:     map[string]string{"text": "write"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(StoreUpper) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(43), result.StateMerges)
	assertStdlibQueryResult(t, engine, states, contract, sender, "GetResult", nil, "WRITE")
}

func TestGnoStdlibMemPackagesAreDeterministicAndContainAllowedRoots(t *testing.T) {
	first, err := GnoStdlibMemPackages()
	if err != nil {
		t.Fatalf("GnoStdlibMemPackages returned error: %v", err)
	}
	second, err := GnoStdlibMemPackages()
	if err != nil {
		t.Fatalf("GnoStdlibMemPackages second call returned error: %v", err)
	}

	firstShape := stdlibPackageShape(first)
	if !reflect.DeepEqual(firstShape, stdlibPackageShape(second)) {
		t.Fatalf("stdlib package order or file ordering is not deterministic")
	}

	paths := map[string]bool{}
	indices := map[string]int{}
	for i, pkg := range first {
		paths[pkg.Path] = true
		indices[pkg.Path] = i
		fileNames := make([]string, 0, len(pkg.Files))
		for _, file := range pkg.Files {
			if strings.HasSuffix(file.Name, "_test.gno") || strings.HasSuffix(file.Name, "_filetest.gno") {
				t.Fatalf("package %q includes test source %q", pkg.Path, file.Name)
			}
			fileNames = append(fileNames, file.Name)
		}
		if !sort.StringsAreSorted(fileNames) {
			t.Fatalf("package %q files are not sorted: %#v", pkg.Path, fileNames)
		}
	}

	for _, pkg := range first {
		deps, err := gnoStdlibImportPaths(pkg)
		if err != nil {
			t.Fatalf("gnoStdlibImportPaths(%q) returned error: %v", pkg.Path, err)
		}
		for _, dep := range deps {
			depIndex, found := indices[dep]
			if !found {
				t.Fatalf("dependency %q of %q was not loaded", dep, pkg.Path)
			}
			if depIndex >= indices[pkg.Path] {
				t.Fatalf("dependency %q must be loaded before %q", dep, pkg.Path)
			}
		}
	}

	for _, importPath := range allowedTypedContractImportPathsByKind(AllowedImportStdlib) {
		if !paths[importPath] {
			t.Fatalf("expected allowed stdlib package %q to be loadable", importPath)
		}
	}
}

func assertStdlibQueryResult(
	t *testing.T,
	engine ContractEngine,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	function string,
	callData map[string]string,
	want string,
) {
	t.Helper()
	if callData == nil {
		callData = map[string]string{}
	}

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: allowedStdlibRuntimeContractSource,
		Function:     function,
		CallData:     callData,
	})
	if err != nil {
		t.Fatalf("QueryContract(%s) returned error: %v", function, err)
	}
	if got, ok := qr.Result.(string); !ok || got != want {
		t.Fatalf("expected %s result %q, got %#v", function, want, qr.Result)
	}
}

func stdlibPackageShape(packages []*gnostd.MemPackage) [][]string {
	out := make([][]string, 0, len(packages))
	for _, pkg := range packages {
		shape := []string{pkg.Path}
		for _, file := range pkg.Files {
			shape = append(shape, file.Name)
		}
		out = append(out, shape)
	}
	return out
}
