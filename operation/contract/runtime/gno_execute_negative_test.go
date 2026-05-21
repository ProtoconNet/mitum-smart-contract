package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const negativeExecuteContractSource = `package contract
import "mitum/chain"

var value string
var limit int64

func Initialize(ctx chain.WriteContext) error {
	return nil
}

func SetValue(ctx chain.WriteContext, next string) error {
	value = next
	return nil
}

func SetLimit(ctx chain.WriteContext, next int64) error {
	limit = next
	return nil
}
`

func TestExecuteContractFunctionNotFoundUsesConsistentCoreWording(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractnege0001")
	sender := base.NewStringAddress("sendernege0001")

	states := registerNegativeExecuteContract(t, engine, contract, sender)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(900),
		ContractCode: negativeExecuteContractSource,
		Function:     "Missing",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected function not found error")
	}
	if !strings.Contains(err.Error(), `function "Missing" not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteContractMissingCallDataArgUsesConsistentCoreWording(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractnege0002")
	sender := base.NewStringAddress("sendernege0002")

	states := registerNegativeExecuteContract(t, engine, contract, sender)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(901),
		ContractCode: negativeExecuteContractSource,
		Function:     "SetValue",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected missing callData arg error")
	}
	if !strings.Contains(err.Error(), `missing callData["next"]`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteContractScalarParseFailureUsesConsistentCoreWording(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractnege0003")
	sender := base.NewStringAddress("sendernege0003")

	states := registerNegativeExecuteContract(t, engine, contract, sender)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(902),
		ContractCode: negativeExecuteContractSource,
		Function:     "SetLimit",
		CallData: map[string]string{
			"next": "not-a-number",
		},
	})
	if err == nil {
		t.Fatal("expected scalar parse failure")
	}
	if !containsAll(err.Error(), `invalid arg "next"`, "invalid syntax") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func registerNegativeExecuteContract(
	t *testing.T,
	engine ContractEngine,
	contract base.Address,
	sender base.Address,
) map[string]base.State {
	t.Helper()

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(899),
		ContractCode: negativeExecuteContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(899), result.StateMerges)
	return states
}
