package contract

import (
	"context"
	"reflect"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	types "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

const registerProcessSchemaReuseContractSource = `package contract

import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
`

type registerProcessSchemaReuseEngine struct {
	t             *testing.T
	expected      cruntime.ContractSchema
	validateCalls int
	executeCalls  int
}

func (e *registerProcessSchemaReuseEngine) ValidateContract(sourceCode string) (cruntime.ContractSchema, base.OperationProcessReasonError) {
	e.validateCalls++
	return e.expected, nil
}

func (e *registerProcessSchemaReuseEngine) ExecuteContract(
	_ encoder.Encoders,
	_ base.GetStateFunc,
	req cruntime.ExecuteRequest,
) (cruntime.ExecuteResult, base.OperationProcessReasonError) {
	e.executeCalls++

	if e.validateCalls != 1 {
		e.t.Fatalf("expected ValidateContract before ExecuteContract, got %d calls", e.validateCalls)
	}
	if req.Schema == nil {
		e.t.Fatal("expected ExecuteRequest.Schema to be set")
	}
	if !reflect.DeepEqual(*req.Schema, e.expected) {
		e.t.Fatalf("unexpected schema passed to ExecuteContract: %#v", *req.Schema)
	}
	if req.ContractCode != registerProcessSchemaReuseContractSource {
		e.t.Fatal("expected ExecuteRequest.ContractCode to be preserved")
	}

	return cruntime.ExecuteResult{
		Engine:      pstate.RuntimeEngineGnoSnapshot,
		StateMerges: nil,
	}, nil
}

func (e *registerProcessSchemaReuseEngine) QueryContract(
	encoder.Encoders,
	base.GetStateFunc,
	cruntime.QueryRequest,
) (cruntime.QueryResult, base.OperationProcessReasonError) {
	e.t.Fatal("QueryContract should not be called in register processor test")
	return cruntime.QueryResult{}, nil
}

func TestRegisterContractProcessorPassesValidatedSchemaToExecute(t *testing.T) {
	originalEngine := contractEngine
	defer func() { contractEngine = originalEngine }()

	expectedSchema := cruntime.ContractSchema{
		PackageName: "contract",
		Mode:        cruntime.SchemaModeTypedArgs,
	}
	fakeEngine := &registerProcessSchemaReuseEngine{
		t:        t,
		expected: expectedSchema,
	}
	contractEngine = fakeEngine

	contractAddr := base.NewStringAddress("contractreg0001")
	sender := base.NewStringAddress("senderreg0001")

	states := map[string]base.State{
		cestate.StateKeyContractAccount(contractAddr): common.NewBaseState(
			base.Height(1),
			cestate.StateKeyContractAccount(contractAddr),
			cestate.NewContractAccountStateValue(types.NewContractAccountStatus(sender, nil)),
			nil,
			nil,
		),
	}
	getStateFunc := func(key string) (base.State, bool, error) {
		st, found := states[key]
		return st, found, nil
	}

	baseProcessor, err := base.NewBaseOperationProcessor(base.Height(33), getStateFunc, nil, nil)
	if err != nil {
		t.Fatalf("NewBaseOperationProcessor returned error: %v", err)
	}

	var encs encoder.Encoders
	opp := &RegisterContractProcessor{
		BaseOperationProcessor: baseProcessor,
		encs:                   &encs,
	}

	fact := NewRegisterContractFact(
		[]byte("token"),
		sender,
		contractAddr,
		registerProcessSchemaReuseContractSource,
		map[string]string{},
		types.CurrencyID("ABC"),
	)
	op, err := NewRegisterContract(fact)
	if err != nil {
		t.Fatalf("NewRegisterContract returned error: %v", err)
	}

	merges, reason, err := opp.Process(context.Background(), op, getStateFunc)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if reason != nil {
		t.Fatalf("Process returned reason: %v", reason)
	}
	if len(merges) != 2 {
		t.Fatalf("expected design + contract account merges, got %d", len(merges))
	}
	if fakeEngine.validateCalls != 1 {
		t.Fatalf("expected ValidateContract to be called once, got %d", fakeEngine.validateCalls)
	}
	if fakeEngine.executeCalls != 1 {
		t.Fatalf("expected ExecuteContract to be called once, got %d", fakeEngine.executeCalls)
	}
}
