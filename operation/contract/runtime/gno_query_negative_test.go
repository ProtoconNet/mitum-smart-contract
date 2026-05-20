package runtime

import (
	"strings"
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
)

const negativeQueryContractSource = `package contract
import "mitum/chain"

var value string

func Initialize(ctx chain.ContractContext) error {
	value = "hello"
	return nil
}

func GetValue(ctx chain.ContractContext) string {
	return value
}

func GetByIndex(ctx chain.ContractContext, index int64) string {
	return value
}
`

func TestQueryContractRuntimeStateMissingStopsBeforeMachineInit(t *testing.T) {
	original := newGnoMachineAndPackageFunc
	machineCalls := 0
	newGnoMachineAndPackageFunc = func(
		*ExecutionContext,
		string,
		string,
		GnoExecutionLimits,
		gstore.GasMeter,
	) (*gno.Machine, *gno.PackageValue, error) {
		machineCalls++
		return nil, nil, nil
	}
	defer func() { newGnoMachineAndPackageFunc = original }()

	engine := NewGnoEngine()
	schema, err := engine.ValidateContract(negativeQueryContractSource)
	if err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}

	_, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), QueryRequest{
		Contract:     base.NewStringAddress("contractnegq0001"),
		Sender:       base.NewStringAddress("sendernegq0001"),
		Height:       base.Height(800),
		ContractCode: negativeQueryContractSource,
		Schema:       &schema,
		Function:     "GetValue",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected runtime state missing error")
	}
	if !strings.Contains(err.Error(), "runtime state not found for typed contract") {
		t.Fatalf("unexpected error: %v", err)
	}
	if machineCalls != 0 {
		t.Fatalf("expected machine init not to run, got %d calls", machineCalls)
	}
}

func TestQueryContractSnapshotStateMissingStopsBeforeMachineInit(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractnegq0002")
	sender := base.NewStringAddress("sendernegq0002")
	schema, err := engine.ValidateContract(negativeQueryContractSource)
	if err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(801),
		ContractCode: negativeQueryContractSource,
		Schema:       &schema,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(801), result.StateMerges)
	delete(states, pstate.SnapshotStateKey(contract))

	original := newGnoMachineAndPackageFunc
	machineCalls := 0
	newGnoMachineAndPackageFunc = func(
		*ExecutionContext,
		string,
		string,
		GnoExecutionLimits,
		gstore.GasMeter,
	) (*gno.Machine, *gno.PackageValue, error) {
		machineCalls++
		return nil, nil, nil
	}
	defer func() { newGnoMachineAndPackageFunc = original }()

	_, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(802),
		ContractCode: negativeQueryContractSource,
		Schema:       &schema,
		Function:     "GetValue",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected snapshot state missing error")
	}
	if !strings.Contains(err.Error(), "snapshot state not found for typed contract") {
		t.Fatalf("unexpected error: %v", err)
	}
	if machineCalls != 0 {
		t.Fatalf("expected machine init not to run, got %d calls", machineCalls)
	}
}

func TestQueryContractFunctionNotFoundStopsBeforeStateLookup(t *testing.T) {
	engine := NewGnoEngine()
	schema, err := engine.ValidateContract(negativeQueryContractSource)
	if err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}

	stateReads := 0
	getStateFunc := func(string) (base.State, bool, error) {
		stateReads++
		return nil, false, nil
	}

	_, err = engine.QueryContract(newRuntimeTestEncoders(t), getStateFunc, QueryRequest{
		Contract:     base.NewStringAddress("contractnegq0003"),
		Sender:       base.NewStringAddress("sendernegq0003"),
		Height:       base.Height(803),
		ContractCode: negativeQueryContractSource,
		Schema:       &schema,
		Function:     "Missing",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected query function not found error")
	}
	if !strings.Contains(err.Error(), `query function "Missing" not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if stateReads != 0 {
		t.Fatalf("expected state lookup not to run, got %d reads", stateReads)
	}
}

func TestQueryContractMissingCallDataArgFailsAtInvocation(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractnegq0004")
	sender := base.NewStringAddress("sendernegq0004")

	states := registerNegativeQueryContract(t, engine, contract, sender)

	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(804),
		ContractCode: negativeQueryContractSource,
		Function:     "GetByIndex",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected missing callData arg error")
	}
	if !strings.Contains(err.Error(), `failed to execute query: missing callData["index"]`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQueryContractScalarParseFailureFailsAtInvocation(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractnegq0005")
	sender := base.NewStringAddress("sendernegq0005")

	states := registerNegativeQueryContract(t, engine, contract, sender)

	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(805),
		ContractCode: negativeQueryContractSource,
		Function:     "GetByIndex",
		CallData: map[string]string{
			"index": "not-a-number",
		},
	})
	if err == nil {
		t.Fatal("expected scalar parse failure")
	}
	if !containsAll(err.Error(), `failed to execute query: invalid arg "index"`, "invalid syntax") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func registerNegativeQueryContract(
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
		Height:       base.Height(799),
		ContractCode: negativeQueryContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(799), result.StateMerges)
	return states
}
