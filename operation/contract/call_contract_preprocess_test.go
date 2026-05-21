package contract

import (
	"context"
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	contracttypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

const callPreProcessValidSource = `package contract

import "mitum/chain"

var value string

func Initialize(ctx chain.WriteContext) error {
	value = "initial"
	return nil
}

func Store(ctx chain.WriteContext, next string) error {
	value = next
	return nil
}
`

func TestCallContractPreProcessRejectsMissingDesignState(t *testing.T) {
	_, reason := runCallContractPreProcess(t, nil, map[string]string{"function": "Store"})
	assertPreProcessReasonContains(t, reason, "contract design not found")
}

func TestCallContractPreProcessRejectsCorruptedDesignState(t *testing.T) {
	contract := base.NewStringAddress("callprecontract01")
	states := map[string]base.State{
		pstate.DesignStateKey(contract): common.NewBaseState(
			base.Height(1),
			pstate.DesignStateKey(contract),
			validCallPreProcessRuntimeState(),
			nil,
			nil,
		),
		pstate.RuntimeStateKey(contract): validCallPreProcessRuntimeBaseState(contract),
	}

	_, reason := runCallContractPreProcessWithContract(t, contract, states, map[string]string{"function": "Store"})
	assertPreProcessReasonContains(t, reason, "failed to decode design state")
}

func TestCallContractPreProcessRejectsMissingRuntimeState(t *testing.T) {
	contract := base.NewStringAddress("callprecontract02")
	states := map[string]base.State{
		pstate.DesignStateKey(contract): validCallPreProcessDesignBaseState(contract, callPreProcessValidSource),
	}

	_, reason := runCallContractPreProcessWithContract(t, contract, states, map[string]string{"function": "Store"})
	assertPreProcessReasonContains(t, reason, "runtime state not found for typed contract")
}

func TestCallContractPreProcessRejectsCorruptedRuntimeState(t *testing.T) {
	contract := base.NewStringAddress("callprecontract03")
	states := map[string]base.State{
		pstate.DesignStateKey(contract): validCallPreProcessDesignBaseState(contract, callPreProcessValidSource),
		pstate.RuntimeStateKey(contract): common.NewBaseState(
			base.Height(1),
			pstate.RuntimeStateKey(contract),
			validCallPreProcessDesignValue(callPreProcessValidSource),
			nil,
			nil,
		),
	}

	_, reason := runCallContractPreProcessWithContract(t, contract, states, map[string]string{"function": "Store"})
	assertPreProcessReasonContains(t, reason, "failed to decode runtime state")
}

func TestCallContractPreProcessRejectsUnsupportedRuntimeEngine(t *testing.T) {
	contract := base.NewStringAddress("callprecontract04")
	states := map[string]base.State{
		pstate.DesignStateKey(contract): validCallPreProcessDesignBaseState(contract, callPreProcessValidSource),
		pstate.RuntimeStateKey(contract): common.NewBaseState(
			base.Height(1),
			pstate.RuntimeStateKey(contract),
			pstate.NewRuntimeStateValue(
				pstate.RuntimeEngine("unsupported-engine-v0"),
				string(cruntime.SchemaModeTypedArgs),
				"contract",
				"mitum.local/r/callpre",
				cruntime.GnoSnapshotVersion,
			),
			nil,
			nil,
		),
	}

	_, reason := runCallContractPreProcessWithContract(t, contract, states, map[string]string{"function": "Store"})
	assertPreProcessReasonContains(t, reason, "unsupported runtime engine")
}

func TestCallContractPreProcessValidCallPasses(t *testing.T) {
	ctx, reason := runCallContractPreProcess(t, validCallPreProcessStates(t, callPreProcessValidSource), map[string]string{
		"function": "Store",
		"next":     "updated",
	})
	if reason != nil {
		t.Fatalf("expected PreProcess to pass, got %v", reason)
	}
	if ctx == nil {
		t.Fatal("expected context to be returned")
	}
}

func TestCallContractPreProcessDoesNotPerformHeavyWork(t *testing.T) {
	originalEngine := contractEngine
	defer func() { contractEngine = originalEngine }()
	contractEngine = callPreProcessNoHeavyWorkEngine{t: t}

	states := validCallPreProcessStates(t, "not valid gno source {{{")
	_, reason := runCallContractPreProcess(t, states, map[string]string{
		"next": "updated",
	})
	if reason != nil {
		t.Fatalf("expected PreProcess to avoid function/schema/runtime execution checks, got %v", reason)
	}
}

type callPreProcessNoHeavyWorkEngine struct {
	t *testing.T
}

func (e callPreProcessNoHeavyWorkEngine) ValidateContract(string) (cruntime.ContractSchema, base.OperationProcessReasonError) {
	e.t.Fatal("PreProcess must not validate or analyze contract schema")
	return cruntime.ContractSchema{}, nil
}

func (e callPreProcessNoHeavyWorkEngine) ExecuteContract(
	encoder.Encoders,
	base.GetStateFunc,
	cruntime.ExecuteRequest,
) (cruntime.ExecuteResult, base.OperationProcessReasonError) {
	e.t.Fatal("PreProcess must not execute contract runtime")
	return cruntime.ExecuteResult{}, nil
}

func (e callPreProcessNoHeavyWorkEngine) QueryContract(
	encoder.Encoders,
	base.GetStateFunc,
	cruntime.QueryRequest,
) (cruntime.QueryResult, base.OperationProcessReasonError) {
	e.t.Fatal("PreProcess must not query contract runtime")
	return cruntime.QueryResult{}, nil
}

func runCallContractPreProcess(
	t *testing.T,
	states map[string]base.State,
	callData map[string]string,
) (context.Context, base.OperationProcessReasonError) {
	t.Helper()

	return runCallContractPreProcessWithContract(
		t,
		base.NewStringAddress("callprecontract00"),
		states,
		callData,
	)
}

func runCallContractPreProcessWithContract(
	t *testing.T,
	contract base.Address,
	states map[string]base.State,
	callData map[string]string,
) (context.Context, base.OperationProcessReasonError) {
	t.Helper()

	if states == nil {
		states = map[string]base.State{}
	}
	sender := base.NewStringAddress("callpresender000")
	getStateFunc := func(key string) (base.State, bool, error) {
		st, found := states[key]
		return st, found, nil
	}

	fact := NewCallContractFact(
		[]byte("call-preprocess-token"),
		sender,
		contract,
		callData,
		currencytypes.CurrencyID("ABC"),
	)
	op, err := NewCallContract(fact)
	if err != nil {
		t.Fatalf("NewCallContract returned error: %v", err)
	}

	opp := &CallContractProcessor{}
	ctx, reason, err := opp.PreProcess(context.Background(), op, getStateFunc)
	if err != nil {
		t.Fatalf("PreProcess returned error: %v", err)
	}

	return ctx, reason
}

func validCallPreProcessStates(t *testing.T, source string) map[string]base.State {
	t.Helper()

	contract := base.NewStringAddress("callprecontract00")
	return map[string]base.State{
		pstate.DesignStateKey(contract):  validCallPreProcessDesignBaseState(contract, source),
		pstate.RuntimeStateKey(contract): validCallPreProcessRuntimeBaseState(contract),
	}
}

func validCallPreProcessDesignBaseState(contract base.Address, source string) base.State {
	return common.NewBaseState(
		base.Height(1),
		pstate.DesignStateKey(contract),
		validCallPreProcessDesignValue(source),
		nil,
		nil,
	)
}

func validCallPreProcessDesignValue(source string) pstate.DesignStateValue {
	return pstate.NewDesignStateValue(contracttypes.NewDesign(source))
}

func validCallPreProcessRuntimeBaseState(contract base.Address) base.State {
	return common.NewBaseState(
		base.Height(1),
		pstate.RuntimeStateKey(contract),
		validCallPreProcessRuntimeState(),
		nil,
		nil,
	)
}

func validCallPreProcessRuntimeState() pstate.RuntimeStateValue {
	return pstate.NewRuntimeStateValue(
		pstate.RuntimeEngineGnoSnapshot,
		string(cruntime.SchemaModeTypedArgs),
		"contract",
		"mitum.local/r/callpre",
		cruntime.GnoSnapshotVersion,
	)
}

func assertPreProcessReasonContains(
	t *testing.T,
	reason base.OperationProcessReasonError,
	substring string,
) {
	t.Helper()

	if reason == nil {
		t.Fatalf("expected PreProcess reason containing %q, got nil", substring)
	}
	if !strings.Contains(reason.Error(), substring) {
		t.Fatalf("expected PreProcess reason containing %q, got %q", substring, reason.Error())
	}
}
