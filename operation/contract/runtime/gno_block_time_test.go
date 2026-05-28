package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const writeContextBlockTimeSource = `package contract
import "mitum/chain"

var writeHeight int64
var writeBlockTime int64

func Initialize(ctx chain.WriteContext) error {
	writeHeight = ctx.GetHeight()
	writeBlockTime = ctx.GetBlockTime()
	return nil
}

func StoreContext(ctx chain.WriteContext) error {
	writeHeight = ctx.GetHeight()
	writeBlockTime = ctx.GetBlockTime()
	return nil
}

func GetWriteHeight(ctx chain.QueryContext) int64 { return writeHeight }
func GetWriteBlockTime(ctx chain.QueryContext) int64 { return writeBlockTime }
`

const queryContextBlockTimeSource = `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func GetBlockTime(ctx chain.QueryContext) int64 { return ctx.GetBlockTime() }
`

func TestWriteContextBlockTimeFollowsExecuteRequest(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractblocktime1")
	sender := base.NewStringAddress("senderblocktime01")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(10),
		BlockTime:    1710000001,
		ContractCode: writeContextBlockTimeSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(10), result.StateMerges)

	assertStoredWriteContextValue(t, engine, states, contract, sender, "GetWriteHeight", 10)
	assertStoredWriteContextValue(t, engine, states, contract, sender, "GetWriteBlockTime", 1710000001)

	result, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(11),
		BlockTime:    1710000999,
		ContractCode: writeContextBlockTimeSource,
		Function:     "StoreContext",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(StoreContext) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(11), result.StateMerges)

	assertStoredWriteContextValue(t, engine, states, contract, sender, "GetWriteHeight", 11)
	assertStoredWriteContextValue(t, engine, states, contract, sender, "GetWriteBlockTime", 1710000999)
}

func TestWriteContextBlockTimeIsDeterministicRequestInput(t *testing.T) {
	const blockTime int64 = 1712345678

	for _, contractID := range []string{"contractblocktime2", "contractblocktime3"} {
		engine := NewGnoEngine()
		contract := base.NewStringAddress(contractID)
		sender := base.NewStringAddress("senderblocktime02")
		states := map[string]base.State{}

		result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
			Mode:         InvocationModeRegister,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(20),
			BlockTime:    blockTime,
			ContractCode: writeContextBlockTimeSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		})
		if err != nil {
			t.Fatalf("ExecuteContract(%s) returned error: %v", contractID, err)
		}
		applyStateMerges(states, base.Height(20), result.StateMerges)
		assertStoredWriteContextValue(t, engine, states, contract, sender, "GetWriteBlockTime", blockTime)
	}
}

func TestQueryContextGetBlockTimeRejected(t *testing.T) {
	_, err := AnalyzeContractSchema(queryContextBlockTimeSource)
	if err == nil || !strings.Contains(err.Error(), "QueryContext.GetBlockTime") {
		t.Fatalf("expected QueryContext.GetBlockTime rejection, got %v", err)
	}
}

func assertStoredWriteContextValue(
	t *testing.T,
	engine ContractEngine,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	function string,
	want int64,
) {
	t.Helper()

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: writeContextBlockTimeSource,
		Function:     function,
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(%s) returned error: %v", function, err)
	}
	if got, ok := qr.Result.(int64); !ok || got != want {
		t.Fatalf("expected %s result %d, got %#v", function, want, qr.Result)
	}
}
