package runtime

import (
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/base"
	pstate "github.com/imfact-labs/smart-contract-model/state"
)

const callerContextContractSource = `package contract
import "mitum/chain"

var lastSender string
var lastCaller string

func Initialize(ctx chain.WriteContext) error {
	lastSender = ""
	lastCaller = ""
	return nil
}

func Record(ctx chain.WriteContext) error {
	lastSender = ctx.GetSender()
	lastCaller = ctx.GetCaller()
	return nil
}

func CallRecord(ctx chain.WriteContext, target string) error {
	return chain.CallContract(ctx, target, "Record", map[string]string{})
}

func ForwardRecord(ctx chain.WriteContext, target string, next string) error {
	return chain.CallContract(ctx, target, "CallRecord", map[string]string{"target": next})
}

func CallRecordAndRecord(ctx chain.WriteContext, target string) error {
	if err := chain.CallContract(ctx, target, "Record", map[string]string{}); err != nil {
		return err
	}
	return Record(ctx)
}

func GetLastSender(ctx chain.QueryContext) string { return lastSender }
func GetLastCaller(ctx chain.QueryContext) string { return lastCaller }
`

func TestWriteContextGetCallerTopLevelEqualsSender(t *testing.T) {
	states, caller, _, _, sender := prepareCallerRuntimeState(t)

	result, err := executeNestedCall(t, states, caller, sender, "Record", map[string]string{})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertCallerContext(t, states, caller, sender, sender.String(), sender.String())
}

func TestWriteContextGetCallerNestedImmediateCaller(t *testing.T) {
	states, caller, target, _, sender := prepareCallerRuntimeState(t)

	result, err := executeNestedCall(t, states, caller, sender, "CallRecord", map[string]string{
		"target": target.String(),
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertCallerContext(t, states, target, sender, sender.String(), caller.String())
}

func TestWriteContextGetCallerNestedChainUsesPreviousContract(t *testing.T) {
	states, caller, middle, leaf, sender := prepareCallerRuntimeState(t)

	result, err := executeNestedCall(t, states, caller, sender, "ForwardRecord", map[string]string{
		"target": middle.String(),
		"next":   leaf.String(),
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertCallerContext(t, states, leaf, sender, sender.String(), middle.String())
}

func TestWriteContextGetSenderOriginSurvivesNestedCall(t *testing.T) {
	states, caller, target, _, sender := prepareCallerRuntimeState(t)

	result, err := executeNestedCall(t, states, caller, sender, "CallRecordAndRecord", map[string]string{
		"target": target.String(),
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertCallerContext(t, states, target, sender, sender.String(), caller.String())
	assertCallerContext(t, states, caller, sender, sender.String(), sender.String())
}

func TestQueryContextGetCallerUnavailable(t *testing.T) {
	states, caller, _, _, sender := prepareCallerRuntimeState(t)
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func Get(ctx chain.QueryContext) string { return ctx.GetCaller() }
`

	_, err := NewGnoEngine().QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     caller,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(caller)].Height(),
		ContractCode: source,
		Function:     "Get",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected QueryContext.GetCaller rejection")
	}
	if strings.Contains(mitumChainPackageSource, "func (ctx QueryContext) GetCaller") {
		t.Fatal("QueryContext must not expose GetCaller")
	}
}

func TestRegisterWriteContextCallerEqualsSender(t *testing.T) {
	states := map[string]base.State{}
	contract := base.NewStringAddress("callerregister001")
	sender := base.NewStringAddress("callerregsender01")
	source := `package contract
import "mitum/chain"
var sender string
var caller string
func Initialize(ctx chain.WriteContext) error { sender=ctx.GetSender();caller=ctx.GetCaller();return nil }
func GetSender(ctx chain.QueryContext) string { return sender }
func GetCaller(ctx chain.QueryContext) string { return caller }
`
	registerNestedRuntimeContract(t, states, contract, sender, source, base.Height(1))
	if got := mustQueryContract(t, states, contract, sender, source, "GetSender", nil).Result.(string); got != sender.String() {
		t.Fatalf("register sender=%q, want %q", got, sender)
	}
	if got := mustQueryContract(t, states, contract, sender, source, "GetCaller", nil).Result.(string); got != sender.String() {
		t.Fatalf("register caller=%q, want %q", got, sender)
	}
}

func TestNestedCallerOriginUint256Integration(t *testing.T) {
	states := map[string]base.State{}
	sender := base.NewStringAddress("calleru256sender1")
	caller := base.NewStringAddress("calleru256caller1")
	target := base.NewStringAddress("calleru256target1")
	source := `package contract
import ("mitum/chain";"mitum/math/v1/u256")
var sender string
var caller string
var result string
func Initialize(ctx chain.WriteContext) error { sender="";caller="";result="0";return nil }
func CallCompute(ctx chain.WriteContext,target string) error { return chain.CallContract(ctx,target,"Compute",map[string]string{}) }
func Compute(ctx chain.WriteContext) error { sender=ctx.GetSender();caller=ctx.GetCaller();v,err:=u256.MulDivCanonicalDecimal("6","7","1");if err!=nil{return err};result=v;return nil }
func GetSender(ctx chain.QueryContext) string { return sender }
func GetCaller(ctx chain.QueryContext) string { return caller }
func GetResult(ctx chain.QueryContext) string { return result }
`
	registerNestedRuntimeContract(t, states, caller, sender, source, base.Height(1))
	registerNestedRuntimeContract(t, states, target, sender, source, base.Height(2))
	r, err := executeNestedCall(t, states, caller, sender, "CallCompute", map[string]string{"target": target.String()})
	if err != nil {
		t.Fatal(err)
	}
	applyStateMerges(states, base.Height(100), r.StateMerges)
	for fn, want := range map[string]string{"GetSender": sender.String(), "GetCaller": caller.String(), "GetResult": "42"} {
		if got := mustQueryContract(t, states, target, sender, source, fn, nil).Result.(string); got != want {
			t.Fatalf("%s=%q, want %q", fn, got, want)
		}
	}
}

func prepareCallerRuntimeState(
	t *testing.T,
) (map[string]base.State, base.Address, base.Address, base.Address, base.Address) {
	t.Helper()

	states := map[string]base.State{}
	sender := base.NewStringAddress("callersender0001")
	caller := base.NewStringAddress("callercaller0001")
	middle := base.NewStringAddress("callermiddle0001")
	leaf := base.NewStringAddress("callerleaf00001")
	registerNestedRuntimeContract(t, states, caller, sender, callerContextContractSource, base.Height(1))
	registerNestedRuntimeContract(t, states, middle, sender, callerContextContractSource, base.Height(2))
	registerNestedRuntimeContract(t, states, leaf, sender, callerContextContractSource, base.Height(3))

	return states, caller, middle, leaf, sender
}

func assertCallerContext(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	wantSender string,
	wantCaller string,
) {
	t.Helper()

	qr := mustQueryContract(t, states, contract, sender, callerContextContractSource, "GetLastSender", map[string]string{})
	if got := qr.Result.(string); got != wantSender {
		t.Fatalf("unexpected sender for %s: got %q, want %q", contract, got, wantSender)
	}
	qr = mustQueryContract(t, states, contract, sender, callerContextContractSource, "GetLastCaller", map[string]string{})
	if got := qr.Result.(string); got != wantCaller {
		t.Fatalf("unexpected caller for %s: got %q, want %q", contract, got, wantCaller)
	}
}
