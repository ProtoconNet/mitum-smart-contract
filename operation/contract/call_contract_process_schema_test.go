package contract

import (
	"context"
	"reflect"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	types "github.com/ProtoconNet/mitum-currency/v3/types"
	ptypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

const callProcessPersistedSchemaSource = `package contract

import "mitum/chain"

var value string

func Initialize(ctx chain.ContractContext) error {
	value = "initial"
	return nil
}

func Store(ctx chain.ContractContext, next string) error {
	value = next
	return nil
}

func GetValue(ctx chain.ContractContext) string { return value }
`

type callProcessSchemaCaptureEngine struct {
	t            *testing.T
	expected     *cruntime.ContractSchema
	executeCalls int
}

func (e *callProcessSchemaCaptureEngine) ValidateContract(string) (cruntime.ContractSchema, base.OperationProcessReasonError) {
	e.t.Fatal("ValidateContract should not be called in call processor schema wiring test")
	return cruntime.ContractSchema{}, nil
}

func (e *callProcessSchemaCaptureEngine) ExecuteContract(
	_ encoder.Encoders,
	_ base.GetStateFunc,
	req cruntime.ExecuteRequest,
) (cruntime.ExecuteResult, base.OperationProcessReasonError) {
	e.executeCalls++

	if e.expected == nil {
		if req.Schema != nil {
			e.t.Fatalf("expected ExecuteRequest.Schema to be nil, got %#v", *req.Schema)
		}
	} else {
		if req.Schema == nil {
			e.t.Fatal("expected ExecuteRequest.Schema to be set")
		}
		if !reflect.DeepEqual(*req.Schema, *e.expected) {
			e.t.Fatalf("unexpected schema passed to ExecuteContract\ngot:  %#v\nwant: %#v", *req.Schema, *e.expected)
		}
	}
	if req.ContractCode != callProcessPersistedSchemaSource {
		e.t.Fatal("expected ExecuteRequest.ContractCode to be preserved")
	}
	if req.Function != "Store" {
		e.t.Fatalf("unexpected function: %q", req.Function)
	}

	return cruntime.ExecuteResult{
		Engine:      pstate.RuntimeEngineGnoSnapshot,
		StateMerges: nil,
	}, nil
}

func (e *callProcessSchemaCaptureEngine) QueryContract(
	encoder.Encoders,
	base.GetStateFunc,
	cruntime.QueryRequest,
) (cruntime.QueryResult, base.OperationProcessReasonError) {
	e.t.Fatal("QueryContract should not be called in call processor schema wiring test")
	return cruntime.QueryResult{}, nil
}

func TestCallContractProcessorPassesPersistedSchemaToExecute(t *testing.T) {
	schema := mustAnalyzeCallProcessSchema(t)
	persisted := cruntime.NewPersistedContractSchema(callProcessPersistedSchemaSource, schema)
	expected, ok := cruntime.RuntimeSchemaFromPersisted(callProcessPersistedSchemaSource, &persisted)
	if !ok {
		t.Fatal("expected persisted schema to be reusable")
	}

	fakeEngine := &callProcessSchemaCaptureEngine{t: t, expected: &expected}
	runCallProcessSchemaTest(t, &persisted, fakeEngine)

	if fakeEngine.executeCalls != 1 {
		t.Fatalf("expected ExecuteContract to be called once, got %d", fakeEngine.executeCalls)
	}
}

func TestCallContractProcessorFallsBackWhenPersistedSchemaMismatches(t *testing.T) {
	schema := mustAnalyzeCallProcessSchema(t)
	persisted := cruntime.NewPersistedContractSchema(callProcessPersistedSchemaSource, schema)
	persisted.SchemaRulesetVersion = "typed-gno-ruleset-v0"

	fakeEngine := &callProcessSchemaCaptureEngine{t: t, expected: nil}
	runCallProcessSchemaTest(t, &persisted, fakeEngine)

	if fakeEngine.executeCalls != 1 {
		t.Fatalf("expected ExecuteContract to be called once, got %d", fakeEngine.executeCalls)
	}
}

func TestCallContractProcessorFallsBackWhenPersistedSchemaMissing(t *testing.T) {
	fakeEngine := &callProcessSchemaCaptureEngine{t: t, expected: nil}
	runCallProcessSchemaTest(t, nil, fakeEngine)

	if fakeEngine.executeCalls != 1 {
		t.Fatalf("expected ExecuteContract to be called once, got %d", fakeEngine.executeCalls)
	}
}

func mustAnalyzeCallProcessSchema(t *testing.T) cruntime.ContractSchema {
	t.Helper()

	schema, err := cruntime.AnalyzeContractSchema(callProcessPersistedSchemaSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	return schema
}

func runCallProcessSchemaTest(
	t *testing.T,
	persisted *ptypes.PersistedContractSchema,
	fakeEngine *callProcessSchemaCaptureEngine,
) {
	t.Helper()

	originalEngine := contractEngine
	defer func() { contractEngine = originalEngine }()
	contractEngine = fakeEngine

	contractAddr := base.NewStringAddress("contractcallmeta1")
	sender := base.NewStringAddress("sendercallmeta01")
	states := map[string]base.State{
		pstate.DesignStateKey(contractAddr): common.NewBaseState(
			base.Height(1),
			pstate.DesignStateKey(contractAddr),
			pstate.NewDesignStateValueWithSchema(
				ptypes.NewDesign(callProcessPersistedSchemaSource),
				persisted,
			),
			nil,
			nil,
		),
	}
	getStateFunc := func(key string) (base.State, bool, error) {
		st, found := states[key]
		return st, found, nil
	}

	baseProcessor, err := base.NewBaseOperationProcessor(base.Height(44), getStateFunc, nil, nil)
	if err != nil {
		t.Fatalf("NewBaseOperationProcessor returned error: %v", err)
	}

	var encs encoder.Encoders
	opp := &CallContractProcessor{
		BaseOperationProcessor: baseProcessor,
		encs:                   &encs,
	}

	fact := NewCallContractFact(
		[]byte("token"),
		sender,
		contractAddr,
		map[string]string{
			"function": "Store",
			"next":     "updated",
		},
		types.CurrencyID("ABC"),
	)
	op, err := NewCallContract(fact)
	if err != nil {
		t.Fatalf("NewCallContract returned error: %v", err)
	}

	merges, reason, err := opp.Process(context.Background(), op, getStateFunc)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if reason != nil {
		t.Fatalf("Process returned reason: %v", reason)
	}
	if len(merges) != 0 {
		t.Fatalf("expected fake engine to return no merges, got %d", len(merges))
	}
}
