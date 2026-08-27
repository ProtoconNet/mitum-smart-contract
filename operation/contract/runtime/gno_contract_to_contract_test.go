package runtime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/imfact-labs/currency-model/common"
	"github.com/imfact-labs/mitum2/base"
	pstate "github.com/imfact-labs/smart-contract-model/state"
	stypes "github.com/imfact-labs/smart-contract-model/types"
)

const nestedCounterContractSource = `package contract
import "mitum/chain"

var revision int64

func Initialize(ctx chain.WriteContext) error {
	revision = 0
	return nil
}

func Bump(ctx chain.WriteContext, amount int64) error {
	revision = revision + amount
	return nil
}

func FailAfterBump(ctx chain.WriteContext, amount int64) error {
	revision = revision + amount
	panic("raw nested failure payload")
}

func Spin(ctx chain.WriteContext) error {
	for {
		revision = revision + 1
	}
	return nil
}

func Call(ctx chain.WriteContext, target string, method string, amount string) error {
	return chain.CallContract(ctx, target, method, map[string]string{"amount": amount})
}

func CallEmpty(ctx chain.WriteContext, target string, method string) error {
	return chain.CallContract(ctx, target, method, map[string]string{})
}

func CallInitialize(ctx chain.WriteContext, target string) error {
	return chain.CallContract(ctx, target, "Initialize", map[string]string{})
}

func Self(ctx chain.WriteContext) error {
	return chain.CallContract(ctx, ctx.GetContract(), "Bump", map[string]string{"amount": "1"})
}

func GetRevision(ctx chain.QueryContext) int64 {
	return revision
}
`

func TestContractToContractCallMutatesExistingTarget(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)

	result, err := executeNestedCall(t, states, caller, sender, "Call", map[string]string{
		"target": target.String(),
		"method": "Bump",
		"amount": "7",
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertNestedRevision(t, states, caller, sender, nestedCounterContractSource, 0)
	assertNestedRevision(t, states, target, sender, nestedCounterContractSource, 7)
}

func TestNestedMutationVisibleToLaterNestedCall(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)

	result, err := executeNestedBatch(t, states, caller, sender, []ExecuteCallItem{
		{Function: "Call", CallData: map[string]string{"target": target.String(), "method": "Bump", "amount": "2"}},
		{Function: "Call", CallData: map[string]string{"target": target.String(), "method": "Bump", "amount": "3"}},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertNestedRevision(t, states, target, sender, nestedCounterContractSource, 5)
}

func TestTopLevelBatchItemSeesPriorOverlayMutation(t *testing.T) {
	states, caller, _, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)

	result, err := executeNestedBatch(t, states, caller, sender, []ExecuteCallItem{
		{Function: "Bump", CallData: map[string]string{"amount": "2"}},
		{Function: "Bump", CallData: map[string]string{"amount": "3"}},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertNestedRevision(t, states, caller, sender, nestedCounterContractSource, 5)
}

func TestNestedFailureRollsBackTouchedSnapshots(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	beforeCaller := snapshotBytesForContract(t, states, caller)
	beforeTarget := snapshotBytesForContract(t, states, target)

	result, err := executeNestedBatch(t, states, caller, sender, []ExecuteCallItem{
		{Function: "Call", CallData: map[string]string{"target": target.String(), "method": "Bump", "amount": "2"}},
		{Function: "Call", CallData: map[string]string{"target": target.String(), "method": "FailAfterBump", "amount": "3"}},
	})
	if err == nil {
		t.Fatal("expected nested failure")
	}
	if len(result.StateMerges) != 0 {
		t.Fatalf("expected no state merges on nested failure, got %d", len(result.StateMerges))
	}
	assertSnapshotBytes(t, states, caller, beforeCaller)
	assertSnapshotBytes(t, states, target, beforeTarget)
	if strings.Contains(err.Error(), "raw nested failure payload") {
		t.Fatalf("failure reason leaked raw panic payload: %v", err)
	}
}

func TestNestedOutOfGasRollsBackTouchedSnapshots(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	beforeTarget := snapshotBytesForContract(t, states, target)

	result, err := executeNestedCall(t, states, caller, sender, "CallEmpty", map[string]string{
		"target": target.String(),
		"method": "Spin",
	})
	if err == nil {
		t.Fatal("expected nested out of gas")
	}
	if !strings.Contains(err.Error(), "nested execution out of gas") {
		t.Fatalf("expected nested OOG reason, got %v", err)
	}
	if len(result.StateMerges) != 0 {
		t.Fatalf("expected no state merges on nested OOG, got %d", len(result.StateMerges))
	}
	assertSnapshotBytes(t, states, target, beforeTarget)
}

func TestNestedDepthLimitRejects(t *testing.T) {
	states, caller, _, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	contracts := make([]base.Address, 0, MaxContractCallDepth+1)
	for i := 0; i < MaxContractCallDepth+1; i++ {
		addr := base.NewStringAddress(fmt.Sprintf("depthcontract%03d", i))
		registerNestedRuntimeContract(t, states, addr, sender, nestedCounterContractSource, base.Height(10+i))
		contracts = append(contracts, addr)
	}
	caller = contracts[0]

	req := ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     caller,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: nestedDepthContractSource(contracts),
		Function:     "Start",
		CallData:     map[string]string{},
	}
	depthSource := nestedDepthContractSource(contracts)
	for _, contract := range contracts {
		replaceContractDesign(t, states, contract, depthSource)
	}
	req.ContractCode = depthSource

	_, err := NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), req)
	if err == nil || !strings.Contains(err.Error(), "depth exceeded") {
		t.Fatalf("expected depth exceeded, got %v", err)
	}
}

func TestNestedSelfCallRejects(t *testing.T) {
	states, caller, _, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)

	_, err := executeNestedCall(t, states, caller, sender, "Self", map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "reentrancy rejected") {
		t.Fatalf("expected self-call rejection, got %v", err)
	}
}

func TestNestedReentrancyRejects(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	source := nestedReentrantContractSource(caller, target)
	replaceContractDesign(t, states, caller, source)
	replaceContractDesign(t, states, target, source)

	_, err := NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     caller,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: source,
		Function:     "Start",
		CallData:     map[string]string{},
	})
	if err == nil || !strings.Contains(err.Error(), "reentrancy rejected") {
		t.Fatalf("expected reentrancy rejection, got %v", err)
	}
}

func TestNestedTouchedLimitRejects(t *testing.T) {
	states, caller, _, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	targets := make([]base.Address, 0, MaxTouchedContractsPerOperation)
	for i := 0; i < MaxTouchedContractsPerOperation; i++ {
		addr := base.NewStringAddress(fmt.Sprintf("touchcontract%03d", i))
		registerNestedRuntimeContract(t, states, addr, sender, nestedCounterContractSource, base.Height(20+i))
		targets = append(targets, addr)
	}
	source := nestedTouchedContractSource(targets)
	replaceContractDesign(t, states, caller, source)

	_, err := NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     caller,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: source,
		Function:     "Start",
		CallData:     map[string]string{},
	})
	if err == nil || !strings.Contains(err.Error(), "touched limit exceeded") {
		t.Fatalf("expected touched limit rejection, got %v", err)
	}
}

func TestNestedInitializeRejects(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)

	_, err := executeNestedCall(t, states, caller, sender, "CallInitialize", map[string]string{
		"target": target.String(),
	})
	if err == nil || !strings.Contains(err.Error(), "nested Initialize rejected") {
		t.Fatalf("expected nested Initialize rejection, got %v", err)
	}
}

func TestCallContractFromQueryPathRejected(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func Get(ctx chain.QueryContext, target string) string {
	_ = chain.CallContract(ctx, target, "Bump", map[string]string{})
	return ""
}
`
	_, err := NewGnoEngine().QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     caller,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(caller)].Height(),
		ContractCode: source,
		Function:     "Get",
		CallData:     map[string]string{"target": target.String()},
	})
	if err == nil {
		t.Fatal("expected query path CallContract rejection")
	}
}

func TestCallContractDuringRegisterRejects(t *testing.T) {
	states := map[string]base.State{}
	sender := base.NewStringAddress("nestedregsender1")
	target := base.NewStringAddress("nestedregtarget1")
	registerNestedRuntimeContract(t, states, target, sender, nestedCounterContractSource, base.Height(1))

	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext, target string) error {
	return chain.CallContract(ctx, target, "Bump", map[string]string{"amount":"1"})
}
`
	contract := base.NewStringAddress("nestedregcaller1")
	_, err := NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: source,
		Function:     "Initialize",
		CallData:     map[string]string{"target": target.String()},
		InitData:     map[string]string{"target": target.String()},
	})
	if err == nil || !strings.Contains(err.Error(), "nested call during RegisterContract rejected") {
		t.Fatalf("expected register nested call rejection, got %v", err)
	}
}

func TestSameSessionRegisteredTargetNotVisibleAsNestedTarget(t *testing.T) {
	states, caller, target, sender := prepareNestedRuntimeState(t, nestedCounterContractSource)
	delete(states, pstate.DesignStateKey(target))

	_, err := executeNestedCall(t, states, caller, sender, "Call", map[string]string{
		"target": target.String(),
		"method": "Bump",
		"amount": "1",
	})
	if err == nil || !strings.Contains(err.Error(), "target design not found") {
		t.Fatalf("expected target design not found, got %v", err)
	}
}

func prepareNestedRuntimeState(
	t *testing.T,
	source string,
) (map[string]base.State, base.Address, base.Address, base.Address) {
	t.Helper()

	states := map[string]base.State{}
	sender := base.NewStringAddress("nestedsender0001")
	caller := base.NewStringAddress("nestedcaller0001")
	target := base.NewStringAddress("nestedtarget0001")
	registerNestedRuntimeContract(t, states, caller, sender, source, base.Height(1))
	registerNestedRuntimeContract(t, states, target, sender, source, base.Height(2))

	return states, caller, target, sender
}

func registerNestedRuntimeContract(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	source string,
	height base.Height,
) {
	t.Helper()

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	result, berr := NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       height,
		ContractCode: source,
		Schema:       &schema,
		Function:     "Initialize",
		CallData:     map[string]string{},
		InitData:     map[string]string{},
	})
	if berr != nil {
		t.Fatalf("ExecuteContract register returned error: %v", berr)
	}
	applyStateMerges(states, height, result.StateMerges)
	addDesignState(t, states, contract, source, schema, height)
}

func replaceContractDesign(t *testing.T, states map[string]base.State, contract base.Address, source string) {
	t.Helper()
	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema replacement returned error: %v", err)
	}
	addDesignState(t, states, contract, source, schema, states[pstate.SnapshotStateKey(contract)].Height())
}

func addDesignState(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	source string,
	schema ContractSchema,
	height base.Height,
) {
	t.Helper()
	persisted := NewPersistedContractSchema(source, schema)
	design := stypes.NewDesign(source)
	states[pstate.DesignStateKey(contract)] = common.NewBaseState(
		height,
		pstate.DesignStateKey(contract),
		pstate.NewDesignStateValueWithSchema(design, &persisted),
		nil,
		nil,
	)
}

func executeNestedCall(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	function string,
	callData map[string]string,
) (ExecuteResult, base.OperationProcessReasonError) {
	t.Helper()
	return executeNestedBatch(t, states, contract, sender, []ExecuteCallItem{{Function: function, CallData: callData}})
}

func executeNestedBatch(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	items []ExecuteCallItem,
) (ExecuteResult, base.OperationProcessReasonError) {
	t.Helper()
	designValue, err := pstate.GetDesignStateValueFromState(states[pstate.DesignStateKey(contract)])
	if err != nil {
		t.Fatalf("GetDesignStateValueFromState returned error: %v", err)
	}
	schema, err := AnalyzeContractSchema(designValue.Design.ContractCode())
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	return NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: designValue.Design.ContractCode(),
		Schema:       &schema,
		CallItems:    items,
	})
}

func assertNestedRevision(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	source string,
	want int64,
) {
	t.Helper()
	qr := mustQueryContract(t, states, contract, sender, source, "GetRevision", map[string]string{})
	if got := qr.Result.(int64); got != want {
		t.Fatalf("unexpected revision for %s: got %d, want %d", contract, got, want)
	}
}

func snapshotBytesForContract(t *testing.T, states map[string]base.State, contract base.Address) []byte {
	t.Helper()
	snapshotValue, err := pstate.GetSnapshotFromState(states[pstate.SnapshotStateKey(contract)])
	if err != nil {
		t.Fatalf("GetSnapshotFromState returned error: %v", err)
	}
	return append([]byte(nil), snapshotValue.Snapshot...)
}

func assertSnapshotBytes(t *testing.T, states map[string]base.State, contract base.Address, want []byte) {
	t.Helper()
	got := snapshotBytesForContract(t, states, contract)
	if string(got) != string(want) {
		t.Fatalf("snapshot changed for %s\nwant: %s\ngot:  %s", contract, want, got)
	}
}

func nestedDepthContractSource(contracts []base.Address) string {
	var b strings.Builder
	b.WriteString("package contract\nimport \"mitum/chain\"\nvar revision int64\n")
	b.WriteString("func Initialize(ctx chain.WriteContext) error { return nil }\n")
	b.WriteString("func Bump(ctx chain.WriteContext, amount int64) error { revision += amount; return nil }\n")
	for i := 0; i < len(contracts)-1; i++ {
		fmt.Fprintf(&b, "func F%d(ctx chain.WriteContext) error { return chain.CallContract(ctx, %q, \"F%d\", map[string]string{}) }\n", i, contracts[i+1].String(), i+1)
	}
	last := len(contracts) - 1
	fmt.Fprintf(&b, "func F%d(ctx chain.WriteContext) error { return nil }\n", last)
	b.WriteString("func Start(ctx chain.WriteContext) error { return F0(ctx) }\n")
	b.WriteString("func GetRevision(ctx chain.QueryContext) int64 { return revision }\n")
	return b.String()
}

func nestedTouchedContractSource(targets []base.Address) string {
	var b strings.Builder
	b.WriteString("package contract\nimport \"mitum/chain\"\nvar revision int64\n")
	b.WriteString("func Initialize(ctx chain.WriteContext) error { return nil }\n")
	b.WriteString("func Bump(ctx chain.WriteContext, amount int64) error { revision += amount; return nil }\n")
	b.WriteString("func Start(ctx chain.WriteContext) error {\n")
	for _, target := range targets {
		fmt.Fprintf(&b, "if err := chain.CallContract(ctx, %q, \"Bump\", map[string]string{\"amount\":\"1\"}); err != nil { return err }\n", target.String())
	}
	b.WriteString("return nil\n}\n")
	b.WriteString("func GetRevision(ctx chain.QueryContext) int64 { return revision }\n")
	return b.String()
}

func nestedReentrantContractSource(caller base.Address, target base.Address) string {
	return fmt.Sprintf(`package contract
import "mitum/chain"

var revision int64

func Initialize(ctx chain.WriteContext) error { return nil }
func Bump(ctx chain.WriteContext, amount int64) error { revision += amount; return nil }
func Start(ctx chain.WriteContext) error {
	return chain.CallContract(ctx, %[1]q, "Back", map[string]string{})
}
func Back(ctx chain.WriteContext) error {
	return chain.CallContract(ctx, %[2]q, "Bump", map[string]string{"amount":"1"})
}
func GetRevision(ctx chain.QueryContext) int64 { return revision }
`, target.String(), caller.String())
}
