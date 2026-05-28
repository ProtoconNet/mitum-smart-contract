package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const queryGasLimitContractSource = `package contract
import (
	"mitum/chain"
	"strings"
)

var lastDigest string

func Initialize(ctx chain.WriteContext) error { return nil }

func BurnWrite(ctx chain.WriteContext) error {
	var out string
	for i := 0; i < 1200; i++ {
		out = chain.SHA3Sum256("x")
	}
	lastDigest = out
	return nil
}

func BurnQuery(ctx chain.QueryContext) string {
	var out string
	for i := 0; i < 1200; i++ {
		out = chain.SHA3Sum256("x")
	}
	return out
}

func NormalSHA3(ctx chain.QueryContext) string {
	return chain.SHA3Sum256("abc")
}

func NormalStdlib(ctx chain.QueryContext) string {
	return strings.ToUpper("ok")
}
`

func TestQueryGasLimitIsSeparateAndSmallerThanWrite(t *testing.T) {
	queryLimits := QueryGnoExecutionLimits()
	writeLimits := WriteGnoExecutionLimits()

	if queryLimits.GasLimit >= writeLimits.GasLimit {
		t.Fatalf("expected query gas limit to be smaller than write gas limit: query=%d write=%d", queryLimits.GasLimit, writeLimits.GasLimit)
	}
	if got, want := queryLimits.GasLimit, int64(1_000_000); got != want {
		t.Fatalf("unexpected query gas limit: got %d, want %d", got, want)
	}
}

func TestQueryGasOutOfGasUsesQueryCategory(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractqgas001")
	sender := base.NewStringAddress("senderqgas0001")
	states := registerQueryGasLimitContract(t, engine, contract, sender)

	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: queryGasLimitContractSource,
		Function:     "BurnQuery",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected query out-of-gas failure")
	}

	got := err.Error()
	if !strings.Contains(got, "gno query out of gas") {
		t.Fatalf("expected query out-of-gas category, got %q", got)
	}
	for _, forbidden := range []string{"gno query panicked", "failed to load typed contract package"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("expected query execution out-of-gas, got %q", got)
		}
	}
}

func TestQueryGasAllowsNormalStdlibAndSHA3Queries(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractqgas002")
	sender := base.NewStringAddress("senderqgas0002")
	states := registerQueryGasLimitContract(t, engine, contract, sender)

	assertQueryGasStringResult(t, engine, states, contract, sender, "NormalSHA3", knownSHA3ABC)
	assertQueryGasStringResult(t, engine, states, contract, sender, "NormalStdlib", "OK")
}

func TestQueryGasLimitDoesNotLowerWriteBudget(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractqgas003")
	sender := base.NewStringAddress("senderqgas0003")
	states := registerQueryGasLimitContract(t, engine, contract, sender)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(702),
		ContractCode: queryGasLimitContractSource,
		Function:     "BurnWrite",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("expected same SHA3 loop to fit write gas budget, got %v", err)
	}
	applyStateMerges(states, base.Height(702), result.StateMerges)
}

func registerQueryGasLimitContract(
	t *testing.T,
	engine ContractEngine,
	contract base.Address,
	sender base.Address,
) map[string]base.State {
	t.Helper()

	states := map[string]base.State{}
	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(700),
		ContractCode: queryGasLimitContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(700), result.StateMerges)

	return states
}

func assertQueryGasStringResult(
	t *testing.T,
	engine ContractEngine,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	function string,
	want string,
) {
	t.Helper()

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: queryGasLimitContractSource,
		Function:     function,
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(%s) returned error: %v", function, err)
	}
	if got, ok := qr.Result.(string); !ok || got != want {
		t.Fatalf("expected %s result %q, got %#v", function, want, qr.Result)
	}
}
