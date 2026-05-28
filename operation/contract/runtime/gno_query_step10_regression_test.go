package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const intQueryRegressionContractSource = `package contract
import "mitum/chain"

var count int

func Initialize(ctx chain.WriteContext) error {
	count = 1
	return nil
}

func Add(ctx chain.WriteContext, delta int) error {
	count = count + delta
	return nil
}

func GetCount(ctx chain.QueryContext) int {
	return count
}
`

func TestGnoQueryIntResultPreservesIntType(t *testing.T) {
	_, states, contract, sender, snapshotBefore := prepareQueryTestState(
		t,
		intQueryRegressionContractSource,
		[]ExecuteRequest{
			{
				Mode:         InvocationModeRegister,
				Height:       base.Height(380),
				Function:     "Initialize",
				CallData:     map[string]string{},
				ContractCode: intQueryRegressionContractSource,
			},
			{
				Mode:         InvocationModeCall,
				Height:       base.Height(381),
				Function:     "Add",
				CallData:     map[string]string{"delta": "2"},
				ContractCode: intQueryRegressionContractSource,
			},
		},
	)

	engine := NewGnoEngine()
	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: intQueryRegressionContractSource,
		Function:     "GetCount",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetCount) returned error: %v", err)
	}

	got, ok := qr.Result.(int)
	if !ok {
		t.Fatalf("expected int result type, got %T (%#v)", qr.Result, qr.Result)
	}
	if got != 3 {
		t.Fatalf("unexpected GetCount result: %d", got)
	}

	assertSnapshotStateUnchanged(t, states, contract, snapshotBefore)
}
