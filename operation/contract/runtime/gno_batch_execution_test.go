package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const batchCallContractSource = `package contract
import (
	"errors"
	"mitum/chain"
)

var value string

func Initialize(ctx chain.WriteContext) error {
	value = "init"
	return nil
}

func SetValue(ctx chain.WriteContext, next string) error {
	value = next
	return nil
}

func PrefixValue(ctx chain.WriteContext, prefix string) error {
	value = prefix + value
	return nil
}

func FailAfterSet(ctx chain.WriteContext, next string) error {
	value = next
	return errors.New("boom")
}

func GetValue(ctx chain.QueryContext) string { return value }
`

func TestGnoExecuteContractBatchRunsItemsSequentially(t *testing.T) {
	engine := NewGnoEngine()
	states, contract, sender := prepareBatchRuntimeState(t, engine)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(901),
		ContractCode: batchCallContractSource,
		CallItems: []ExecuteCallItem{
			{Function: "SetValue", CallData: map[string]string{"next": "one"}},
			{Function: "PrefixValue", CallData: map[string]string{"prefix": "two-"}},
		},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(batch) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(901), result.StateMerges)

	qr := mustBatchQueryValue(t, engine, states, contract, sender)
	if qr.Result != "two-one" {
		t.Fatalf("expected second batch item to see first mutation, got %#v", qr.Result)
	}
}

func TestGnoExecuteContractBatchFailureReturnsNoStateMerge(t *testing.T) {
	engine := NewGnoEngine()
	states, contract, sender := prepareBatchRuntimeState(t, engine)
	before := snapshotBytesForBatchTest(t, states, contract)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(902),
		ContractCode: batchCallContractSource,
		CallItems: []ExecuteCallItem{
			{Function: "SetValue", CallData: map[string]string{"next": "temporary"}},
			{Function: "FailAfterSet", CallData: map[string]string{"next": "bad"}},
		},
	})
	if err == nil {
		t.Fatal("expected batch failure")
	}
	if !strings.Contains(err.Error(), `call item 2 "FailAfterSet"`) {
		t.Fatalf("expected item-indexed error, got %v", err)
	}
	if len(result.StateMerges) != 0 {
		t.Fatalf("expected failed batch to return no state merges, got %d", len(result.StateMerges))
	}
	after := snapshotBytesForBatchTest(t, states, contract)
	if string(before) != string(after) {
		t.Fatalf("expected failed batch to leave stored snapshot unchanged\nbefore: %s\nafter:  %s", before, after)
	}

	qr := mustBatchQueryValue(t, engine, states, contract, sender)
	if qr.Result != "init" {
		t.Fatalf("expected failed batch mutation to be rolled back, got %#v", qr.Result)
	}
}

func TestGnoExecuteContractBatchRejectsInitializeItem(t *testing.T) {
	engine := NewGnoEngine()
	states, contract, sender := prepareBatchRuntimeState(t, engine)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(903),
		ContractCode: batchCallContractSource,
		CallItems: []ExecuteCallItem{
			{Function: "Initialize", CallData: map[string]string{}},
		},
	})
	if err == nil {
		t.Fatal("expected Initialize batch item to fail")
	}
	if !strings.Contains(err.Error(), `call item 1 "Initialize"`) {
		t.Fatalf("expected item-indexed Initialize error, got %v", err)
	}
}

func prepareBatchRuntimeState(
	t *testing.T,
	engine ContractEngine,
) (map[string]base.State, base.Address, base.Address) {
	t.Helper()

	contract := base.NewStringAddress("batchrtcontract01")
	sender := base.NewStringAddress("batchrtsender001")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(900),
		ContractCode: batchCallContractSource,
		InitData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(register) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(900), result.StateMerges)

	return states, contract, sender
}

func mustBatchQueryValue(
	t *testing.T,
	engine ContractEngine,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
) QueryResult {
	t.Helper()

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        base.Height(904),
		CurrentHeight: base.Height(904),
		ContractCode:  batchCallContractSource,
		Function:      "GetValue",
		CallData:      map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetValue) returned error: %v", err)
	}

	return qr
}

func snapshotBytesForBatchTest(t *testing.T, states map[string]base.State, contract base.Address) []byte {
	t.Helper()

	snapshotValue, err := state.GetSnapshotFromState(states[state.SnapshotStateKey(contract)])
	if err != nil {
		t.Fatalf("GetSnapshotFromState returned error: %v", err)
	}

	return append([]byte(nil), snapshotValue.Snapshot...)
}
