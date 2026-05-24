package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const packageLoadBaselineSource = `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func Store(ctx chain.WriteContext, value string) error { return nil }
func Get(ctx chain.QueryContext) string { return "" }
`

const packageLoadBadRegisterSelectorSource = `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error {
	_ = ctx.getHeight()
	return nil
}
`

const packageLoadMissingStdlibSymbolSource = `package contract
import (
	"mitum/chain"
	"strings"
)

func Initialize(ctx chain.WriteContext) error { return nil }
func Store(ctx chain.WriteContext, value string) error {
	_ = strings.NotAFunction(value)
	return nil
}
func Get(ctx chain.QueryContext) string { return "" }
`

const packageLoadWrongQueryContextMethodSource = `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func Store(ctx chain.WriteContext, value string) error { return nil }
func Get(ctx chain.QueryContext) string { return ctx.GetSender() }
`

func TestRegisterSurfacesTypedContractPackageLoadFailure(t *testing.T) {
	engine := NewGnoEngine()

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     base.NewStringAddress("contractloaderr01"),
		Sender:       base.NewStringAddress("senderloaderr001"),
		Height:       base.Height(1000),
		ContractCode: packageLoadBadRegisterSelectorSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	assertTypedContractPackageLoadError(t, err, "getHeight")
}

func TestCallSurfacesMissingStdlibSymbolAsPackageLoadFailure(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractloaderr02")
	sender := base.NewStringAddress("senderloaderr002")
	states := registerPackageLoadBaseline(t, engine, contract, sender)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(1001),
		ContractCode: packageLoadMissingStdlibSymbolSource,
		Function:     "Store",
		CallData:     map[string]string{"value": "value"},
	})
	assertTypedContractPackageLoadError(t, err, "NotAFunction")
}

func TestQuerySurfacesWrongContextMethodAsPackageLoadFailure(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractloaderr03")
	sender := base.NewStringAddress("senderloaderr003")
	states := registerPackageLoadBaseline(t, engine, contract, sender)

	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(1002),
		ContractCode: packageLoadWrongQueryContextMethodSource,
		Function:     "Get",
		CallData:     map[string]string{},
	})
	assertTypedContractPackageLoadError(t, err, "GetSender")
}

func TestSanitizeGnoPackageLoadPanicKeepsDiagnosticOnly(t *testing.T) {
	raw := "contract.gno:7: undefined selector NotAFunction\n--- preprocess stack trace ---\nraw-sensitive-vm-dump"
	got := sanitizeGnoPackageLoadPanic(raw)

	if !strings.Contains(got, "NotAFunction") {
		t.Fatalf("expected compact diagnostic, got %q", got)
	}
	if strings.Contains(got, "raw-sensitive-vm-dump") || strings.Contains(got, "stack trace") || strings.Contains(got, "\n") {
		t.Fatalf("package-load surface leaked VM dump content: %q", got)
	}

	oversized := strings.Repeat("x", maxGnoPackageLoadCauseBytes+50)
	if got := sanitizeGnoPackageLoadPanic(oversized); len(got) > maxGnoPackageLoadCauseBytes {
		t.Fatalf("expected package-load diagnostic to be bounded, got %d bytes", len(got))
	}
}

func TestPackageLoadResourceLimitPanicsStayInExecutionLimitClassifier(t *testing.T) {
	gasMeter := NewGnoGasMeter(1)
	gasMeter.ConsumeGas(1, "test")
	if !isGnoPackageLoadResourceLimitPanic("compiler detail", gasMeter) {
		t.Fatal("expected exhausted gas meter to remain a resource-limit panic")
	}
	if !isGnoPackageLoadResourceLimitPanic("allocator limit reached", nil) {
		t.Fatal("expected allocation panic to remain a resource-limit panic")
	}
	if isGnoPackageLoadResourceLimitPanic("undefined selector NotAFunction", nil) {
		t.Fatal("expected ordinary typecheck failure to use package-load diagnostics")
	}
}

func registerPackageLoadBaseline(
	t *testing.T,
	engine ContractEngine,
	contract base.Address,
	sender base.Address,
) map[string]base.State {
	t.Helper()

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(999),
		ContractCode: packageLoadBaselineSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract baseline register returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(999), result.StateMerges)
	return states
}

func assertTypedContractPackageLoadError(t *testing.T, err base.OperationProcessReasonError, cause string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected typed contract package load failure")
	}
	got := err.Error()
	if !strings.Contains(got, "failed to load typed contract package") || !strings.Contains(got, cause) {
		t.Fatalf("expected typed contract package load diagnostic containing %q, got %q", cause, got)
	}
	if strings.Contains(got, "gno execution panicked") || strings.Contains(got, "gno query panicked") || strings.Contains(got, "\n") {
		t.Fatalf("expected package load failure instead of generic or multi-line panic surface, got %q", got)
	}
}
