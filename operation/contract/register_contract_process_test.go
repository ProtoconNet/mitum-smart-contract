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
	contracttypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

const registerProcessSchemaReuseContractSource = `package contract

import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
`

type registerProcessSchemaReuseEngine struct {
	t             *testing.T
	expected      cruntime.ContractSchema
	expectedTime  int64
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
	if req.BlockTime != e.expectedTime {
		e.t.Fatalf("unexpected block time passed to ExecuteContract: %d", req.BlockTime)
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

	expectedSchema, err := cruntime.AnalyzeContractSchema(registerProcessSchemaReuseContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	contractAddr := base.NewStringAddress("contractreg0001")
	sender := base.NewStringAddress("senderreg0001")
	proposalFact := isaac.NewProposalFact(base.GenesisPoint, sender, nil, nil)
	var proposal base.ProposalSignFact = isaac.NewProposalSignFact(proposalFact)
	fakeEngine := &registerProcessSchemaReuseEngine{
		t:            t,
		expected:     expectedSchema,
		expectedTime: proposal.ProposalFact().ProposedAt().Unix(),
	}
	contractEngine = fakeEngine

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

	var encs encoder.Encoders
	processor, err := NewRegisterContractProcessor(encs)(base.Height(33), &proposal, getStateFunc, nil, nil)
	if err != nil {
		t.Fatalf("NewRegisterContractProcessor returned error: %v", err)
	}
	opp, ok := processor.(*RegisterContractProcessor)
	if !ok {
		t.Fatalf("expected RegisterContractProcessor, got %T", processor)
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
	designValue := designStateValueFromMerges(t, merges, pstate.DesignStateKey(contractAddr))
	if designValue.Schema == nil {
		t.Fatal("expected register to store persisted schema metadata")
	}
	if designValue.Schema.SchemaFormatVersion != contracttypes.CurrentSchemaFormatVersion {
		t.Fatalf("unexpected schema format version: %q", designValue.Schema.SchemaFormatVersion)
	}
	if designValue.Schema.SchemaRulesetVersion != cruntime.CurrentSchemaRulesetVersion {
		t.Fatalf("unexpected schema ruleset version: %q", designValue.Schema.SchemaRulesetVersion)
	}
	if designValue.Schema.SourceHash != contracttypes.ContractSourceHash(registerProcessSchemaReuseContractSource) {
		t.Fatalf("unexpected source hash: %q", designValue.Schema.SourceHash)
	}
	storedSchema, ok := cruntime.RuntimeSchemaFromPersisted(registerProcessSchemaReuseContractSource, designValue.Schema)
	if !ok {
		t.Fatal("stored persisted schema was not reusable")
	}
	if !reflect.DeepEqual(storedSchema, expectedSchema) {
		t.Fatalf("stored schema mismatch\ngot:  %#v\nwant: %#v", storedSchema, expectedSchema)
	}
	if fakeEngine.validateCalls != 1 {
		t.Fatalf("expected ValidateContract to be called once, got %d", fakeEngine.validateCalls)
	}
	if fakeEngine.executeCalls != 1 {
		t.Fatalf("expected ExecuteContract to be called once, got %d", fakeEngine.executeCalls)
	}
}

func designStateValueFromMerges(
	t *testing.T,
	merges []base.StateMergeValue,
	key string,
) pstate.DesignStateValue {
	t.Helper()

	for _, merge := range merges {
		if merge.Key() != key {
			continue
		}

		value, ok := merge.Value().(pstate.DesignStateValue)
		if !ok {
			t.Fatalf("expected DesignStateValue merge, got %T", merge.Value())
		}

		return value
	}

	t.Fatalf("design state merge %q not found", key)
	return pstate.DesignStateValue{}
}
