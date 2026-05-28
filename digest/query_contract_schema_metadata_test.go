package digest

import (
	"reflect"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum-smart-contract/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

const digestPersistedSchemaSource = `package contract

import "mitum/chain"

var value string

func Initialize(ctx chain.WriteContext) error {
	value = "initial"
	return nil
}

func GetValue(ctx chain.QueryContext) string { return value }
`

type digestQuerySchemaCaptureEngine struct {
	t          *testing.T
	expected   *runtime.ContractSchema
	queryCalls int
}

func (e *digestQuerySchemaCaptureEngine) ValidateContract(string) (runtime.ContractSchema, base.OperationProcessReasonError) {
	e.t.Fatal("ValidateContract should not be called in digest query schema wiring test")
	return runtime.ContractSchema{}, nil
}

func (e *digestQuerySchemaCaptureEngine) ExecuteContract(
	encoder.Encoders,
	base.GetStateFunc,
	runtime.ExecuteRequest,
) (runtime.ExecuteResult, base.OperationProcessReasonError) {
	e.t.Fatal("ExecuteContract should not be called in digest query schema wiring test")
	return runtime.ExecuteResult{}, nil
}

func (e *digestQuerySchemaCaptureEngine) QueryContract(
	_ encoder.Encoders,
	_ base.GetStateFunc,
	req runtime.QueryRequest,
) (runtime.QueryResult, base.OperationProcessReasonError) {
	e.queryCalls++

	if e.expected == nil {
		if req.Schema != nil {
			e.t.Fatalf("expected QueryRequest.Schema to be nil, got %#v", *req.Schema)
		}
	} else {
		if req.Schema == nil {
			e.t.Fatal("expected QueryRequest.Schema to be set")
		}
		if !reflect.DeepEqual(*req.Schema, *e.expected) {
			e.t.Fatalf("unexpected schema passed to QueryContract\ngot:  %#v\nwant: %#v", *req.Schema, *e.expected)
		}
	}
	if req.ContractCode != digestPersistedSchemaSource {
		e.t.Fatal("expected QueryRequest.ContractCode to be preserved")
	}
	if req.Function != "GetValue" {
		e.t.Fatalf("unexpected function: %q", req.Function)
	}

	return runtime.QueryResult{
		Engine: state.RuntimeEngineGnoSnapshot,
		Result: "ok",
	}, nil
}

func TestDigestQueryPassesPersistedSchemaToRuntime(t *testing.T) {
	schema := mustAnalyzeDigestPersistedSchema(t)
	persisted := runtime.NewPersistedContractSchema(digestPersistedSchemaSource, schema)
	expected, ok := runtime.RuntimeSchemaFromPersisted(digestPersistedSchemaSource, &persisted)
	if !ok {
		t.Fatal("expected persisted schema to be reusable")
	}

	fakeEngine := &digestQuerySchemaCaptureEngine{t: t, expected: &expected}
	hd, contract, states := newDigestPersistedSchemaHandlers(t, &persisted)
	before := designSchemaBytesFromStates(t, states, contract)

	withDigestQueryEngine(t, fakeEngine, func() {
		status, body, _ := performContractQueryRequest(t, hd, contract.String(), map[string]string{
			"function": "GetValue",
		})
		if status != 200 {
			t.Fatalf("unexpected status: %d body=%s", status, body)
		}
	})

	if fakeEngine.queryCalls != 1 {
		t.Fatalf("expected QueryContract to be called once, got %d", fakeEngine.queryCalls)
	}
	after := designSchemaBytesFromStates(t, states, contract)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("digest query mutated DesignState persisted schema metadata")
	}
}

func TestDigestQueryFallsBackWhenPersistedSchemaMismatches(t *testing.T) {
	schema := mustAnalyzeDigestPersistedSchema(t)
	persisted := runtime.NewPersistedContractSchema(digestPersistedSchemaSource, schema)
	persisted.SchemaFormatVersion = "contract-schema-format-v0"

	fakeEngine := &digestQuerySchemaCaptureEngine{t: t, expected: nil}
	hd, contract, _ := newDigestPersistedSchemaHandlers(t, &persisted)

	withDigestQueryEngine(t, fakeEngine, func() {
		status, body, _ := performContractQueryRequest(t, hd, contract.String(), map[string]string{
			"function": "GetValue",
		})
		if status != 200 {
			t.Fatalf("unexpected status: %d body=%s", status, body)
		}
	})

	if fakeEngine.queryCalls != 1 {
		t.Fatalf("expected QueryContract to be called once, got %d", fakeEngine.queryCalls)
	}
}

func TestDigestQueryFallsBackWhenPersistedSchemaMissing(t *testing.T) {
	fakeEngine := &digestQuerySchemaCaptureEngine{t: t, expected: nil}
	hd, contract, _ := newDigestPersistedSchemaHandlers(t, nil)

	withDigestQueryEngine(t, fakeEngine, func() {
		status, body, _ := performContractQueryRequest(t, hd, contract.String(), map[string]string{
			"function": "GetValue",
		})
		if status != 200 {
			t.Fatalf("unexpected status: %d body=%s", status, body)
		}
	})

	if fakeEngine.queryCalls != 1 {
		t.Fatalf("expected QueryContract to be called once, got %d", fakeEngine.queryCalls)
	}
}

func mustAnalyzeDigestPersistedSchema(t *testing.T) runtime.ContractSchema {
	t.Helper()

	schema, err := runtime.AnalyzeContractSchema(digestPersistedSchemaSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	return schema
}

func withDigestQueryEngine(t *testing.T, engine runtime.ContractEngine, fn func()) {
	t.Helper()

	originalEngine := digestContractQueryEngine
	defer func() { digestContractQueryEngine = originalEngine }()
	digestContractQueryEngine = engine

	fn()
}

func newDigestPersistedSchemaHandlers(
	t *testing.T,
	persisted *types.PersistedContractSchema,
) (*Handlers, base.Address, map[string]base.State) {
	t.Helper()

	encsPtr, enc := newDigestTestEncoders(t)
	contract := base.NewStringAddress("contractqmeta001")
	states := map[string]base.State{
		state.DesignStateKey(contract): common.NewBaseState(
			base.Height(1),
			state.DesignStateKey(contract),
			state.NewDesignStateValueWithSchema(
				types.NewDesign(digestPersistedSchemaSource),
				persisted,
			),
			nil,
			nil,
		),
	}

	return newDigestHandlersForStates(t, encsPtr, enc, states), contract, states
}

func designSchemaBytesFromStates(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
) []byte {
	t.Helper()

	value, err := state.GetDesignStateValueFromState(states[state.DesignStateKey(contract)])
	if err != nil {
		t.Fatalf("GetDesignStateValueFromState returned error: %v", err)
	}
	if value.Schema == nil {
		return nil
	}

	return append([]byte(nil), value.Schema.Bytes()...)
}
