package runtime

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
	"github.com/imfact-labs/mitum2/base"
)

const nestedUint256CallerSource = `package contract
import "mitum/chain"
var stored string
func Initialize(ctx chain.WriteContext) error { stored="caller-initial"; return nil }
func Call(ctx chain.WriteContext,target string,method string,a string,b string,d string) error {
	return chain.CallContract(ctx,target,method,map[string]string{"a":a,"b":b,"d":d})
}
func CallTwo(ctx chain.WriteContext,target string) error {
	if err:=chain.CallContract(ctx,target,"MulDiv",map[string]string{"a":"81","b":"1","d":"1"});err!=nil{return err}
	return chain.CallContract(ctx,target,"SqrtStored",map[string]string{})
}
func CallRelay(ctx chain.WriteContext,target string,next string) error {
	return chain.CallContract(ctx,target,"Relay",map[string]string{"next":next})
}
func AtomicOOG(ctx chain.WriteContext,first string,second string) error {
	stored="caller-changed"
	if err:=chain.CallContract(ctx,first,"MulDiv",map[string]string{"a":"81","b":"1","d":"1"});err!=nil{return err}
	return chain.CallContract(ctx,second,"Burn",map[string]string{"n":"200"})
}
func AtomicFailure(ctx chain.WriteContext,first string,second string) error {
	stored="caller-changed"
	if err:=chain.CallContract(ctx,first,"MulDiv",map[string]string{"a":"81","b":"1","d":"1"});err!=nil{return err}
	return chain.CallContract(ctx,second,"Fail",map[string]string{})
}
func Get(ctx chain.QueryContext) string { return stored }
`

const nestedUint256TargetSource = `package contract
import ("mitum/math/v1/u256";"mitum/chain")
var stored string
func Initialize(ctx chain.WriteContext) error { stored="target-initial"; return nil }
func Canonical(ctx chain.WriteContext,a string,b string,d string) error { v,err:=u256.FromCanonicalDecimal(a);if err!=nil{return err};stored=u256.ToCanonicalDecimal(v);return nil }
func MulDiv(ctx chain.WriteContext,a string,b string,d string) error { v,err:=u256.MulDivCanonicalDecimal(a,b,d);if err!=nil{return err};stored=v;return nil }
func Sqrt(ctx chain.WriteContext,a string,b string,d string) error { v,err:=u256.SqrtCanonicalDecimal(a);if err!=nil{return err};stored=v;return nil }
func SqrtStored(ctx chain.WriteContext) error { v,err:=u256.SqrtCanonicalDecimal(stored);if err!=nil{return err};stored=v;return nil }
func Relay(ctx chain.WriteContext,next string) error { return chain.CallContract(ctx,next,"Canonical",map[string]string{"a":"1","b":"0","d":"1"}) }
func Burn(ctx chain.WriteContext,n int64) error { stored="burning";for i:=int64(0);i<n;i++{v,err:=u256.MulDivCanonicalDecimal("1","1","1");if err!=nil{return err};stored=v};return nil }
func Fail(ctx chain.WriteContext) error { stored="changed-before-failure";_,err:=u256.MulDivCanonicalDecimal("1","1","0");return err }
func Get(ctx chain.QueryContext) string { return stored }
`

type nestedUint256Env struct {
	states                        map[string]base.State
	caller, first, second, sender base.Address
}

func newNestedUint256Env(t *testing.T) *nestedUint256Env {
	t.Helper()
	e := &nestedUint256Env{states: map[string]base.State{}, caller: base.NewStringAddress("nestedu256caller1"), first: base.NewStringAddress("nestedu256target1"), second: base.NewStringAddress("nestedu256target2"), sender: base.NewStringAddress("nestedu256sender1")}
	registerNestedRuntimeContract(t, e.states, e.caller, e.sender, nestedUint256CallerSource, 1)
	registerNestedRuntimeContract(t, e.states, e.first, e.sender, nestedUint256TargetSource, 2)
	registerNestedRuntimeContract(t, e.states, e.second, e.sender, nestedUint256TargetSource, 3)
	return e
}

func (e *nestedUint256Env) execute(t *testing.T, items []ExecuteCallItem) (ExecuteResult, base.OperationProcessReasonError) {
	t.Helper()
	schema, err := AnalyzeContractSchema(nestedUint256CallerSource)
	if err != nil {
		t.Fatal(err)
	}
	return NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.caller, Sender: e.sender, Height: 100, ContractCode: nestedUint256CallerSource, Schema: &schema, CallItems: items})
}

func recordNestedUint256Machines(t *testing.T, run func()) ([]gstore.GasMeter, []GnoExecutionLimits) {
	t.Helper()
	original := newGnoMachineAndPackageFunc
	meters := []gstore.GasMeter{}
	limits := []GnoExecutionLimits{}
	newGnoMachineAndPackageFunc = func(ctx *ExecutionContext, path, source string, l GnoExecutionLimits, meter gstore.GasMeter) (*gno.Machine, *gno.PackageValue, error) {
		meters = append(meters, meter)
		limits = append(limits, l)
		return original(ctx, path, source, l, meter)
	}
	defer func() { newGnoMachineAndPackageFunc = original }()
	run()
	return meters, limits
}

func TestNestedUint256ExactSharedGasMeterIdentity(t *testing.T) {
	e := newNestedUint256Env(t)
	meters, limits := recordNestedUint256Machines(t, func() {
		r, err := e.execute(t, []ExecuteCallItem{{Function: "CallRelay", CallData: map[string]string{"target": e.first.String(), "next": e.second.String()}}})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.StateMerges) != 3 {
			t.Fatalf("merges=%d, want 3", len(r.StateMerges))
		}
	})
	if len(meters) != 3 {
		t.Fatalf("machines=%d, want top-level plus two nested", len(meters))
	}
	for i := range meters {
		if meters[i] != meters[0] {
			t.Fatalf("machine %d received a distinct gas meter", i)
		}
		if limits[i].MaxAllocBytes != GnoWriteMaxAllocBytes {
			t.Fatalf("machine %d alloc limit=%d", i, limits[i].MaxAllocBytes)
		}
	}
}

func TestNestedUint256NativePathsAndOverlayVisibility(t *testing.T) {
	e := newNestedUint256Env(t)
	before := snapshotBytesForContract(t, e.states, e.first)
	items := []ExecuteCallItem{
		{Function: "Call", CallData: map[string]string{"target": e.first.String(), "method": "Canonical", "a": u256Max, "b": "0", "d": "1"}},
		{Function: "Call", CallData: map[string]string{"target": e.second.String(), "method": "Sqrt", "a": u256Max, "b": "0", "d": "1"}},
		{Function: "CallTwo", CallData: map[string]string{"target": e.first.String()}},
	}
	r, err := e.execute(t, items)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, snapshotBytesForContract(t, e.states, e.first)) {
		t.Fatal("base snapshot changed before merges were applied")
	}
	keys := make([]string, len(r.StateMerges))
	for i, m := range r.StateMerges {
		keys[i] = m.Key()
	}
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("state merges are not deterministic: %v", keys)
	}
	applyStateMerges(e.states, 100, r.StateMerges)
	if got := mustQueryContract(t, e.states, e.first, e.sender, nestedUint256TargetSource, "Get", nil).Result.(string); got != "9" {
		t.Fatalf("overlay sqrt result=%q", got)
	}
	if got := mustQueryContract(t, e.states, e.second, e.sender, nestedUint256TargetSource, "Get", nil).Result.(string); got != "340282366920938463463374607431768211455" {
		t.Fatalf("nested sqrt result=%q", got)
	}
}

func TestNestedUint256GasAccumulationDeterministic(t *testing.T) {
	e := newNestedUint256Env(t)
	measure := func(items []ExecuteCallItem) int64 {
		var result ExecuteResult
		var execErr base.OperationProcessReasonError
		meters, _ := recordNestedUint256Machines(t, func() { result, execErr = e.execute(t, items) })
		_ = result
		if execErr != nil {
			t.Fatal(execErr)
		}
		return int64(meters[0].GasConsumed())
	}
	one := []ExecuteCallItem{{Function: "Call", CallData: map[string]string{"target": e.first.String(), "method": "MulDiv", "a": u256Max, "b": u256Max, "d": u256Max}}}
	multi := append(append([]ExecuteCallItem{}, one...), ExecuteCallItem{Function: "Call", CallData: map[string]string{"target": e.second.String(), "method": "Sqrt", "a": u256Max, "b": "0", "d": "1"}})
	oneGas, oneAgain := measure(one), measure(one)
	if oneGas != oneAgain {
		t.Fatalf("single gas changed: %d != %d", oneGas, oneAgain)
	}
	multiGas, multiAgain := measure(multi), measure(multi)
	if multiGas != multiAgain {
		t.Fatalf("multi gas changed: %d != %d", multiGas, multiAgain)
	}
	delta := multiGas - oneGas
	if delta <= mitumNativeCallContractGasBase+uint256NativeSqrtGasBase {
		t.Fatalf("gas delta=%d is below added flat tiers", delta)
	}
	t.Logf("NESTED_UINT256_GAS single=%d multi=%d delta=%d", oneGas, multiGas, delta)
}

func TestNestedUint256OutOfGasAtomicRollback(t *testing.T) {
	e := newNestedUint256Env(t)
	contracts := []base.Address{e.caller, e.first, e.second}
	before := map[string][]byte{}
	for _, c := range contracts {
		before[c.String()] = snapshotBytesForContract(t, e.states, c)
	}
	var result ExecuteResult
	var execErr base.OperationProcessReasonError
	meters, _ := recordNestedUint256Machines(t, func() {
		result, execErr = e.execute(t, []ExecuteCallItem{{Function: "AtomicOOG", CallData: map[string]string{"first": e.first.String(), "second": e.second.String()}}})
	})
	if execErr == nil || !strings.Contains(execErr.Error(), "out of gas") {
		t.Fatalf("expected OOG, got %v", execErr)
	}
	if len(result.StateMerges) != 0 {
		t.Fatalf("OOG merges=%d", len(result.StateMerges))
	}
	for _, c := range contracts {
		assertSnapshotBytes(t, e.states, c, before[c.String()])
	}
	for _, forbidden := range []string{"/Users/", "pkg/mod/", "goroutine ", "gas descriptor", "nativeCallContract"} {
		if strings.Contains(execErr.Error(), forbidden) {
			t.Fatalf("OOG exposed %q: %v", forbidden, execErr)
		}
	}
	t.Logf("NESTED_UINT256_OOG consumed=%d limit=%d", meters[0].GasConsumed(), meters[0].Limit())
}

func TestNestedUint256NativeFailureAtomicRollback(t *testing.T) {
	e := newNestedUint256Env(t)
	contracts := []base.Address{e.caller, e.first, e.second}
	before := map[string][]byte{}
	for _, c := range contracts {
		before[c.String()] = snapshotBytesForContract(t, e.states, c)
	}
	r, err := e.execute(t, []ExecuteCallItem{{Function: "AtomicFailure", CallData: map[string]string{"first": e.first.String(), "second": e.second.String()}}})
	if err == nil || !strings.Contains(err.Error(), Uint256MulDivDenominatorZero) {
		t.Fatalf("failure reason=%v", err)
	}
	if len(r.StateMerges) != 0 {
		t.Fatalf("failure merges=%d", len(r.StateMerges))
	}
	for _, c := range contracts {
		assertSnapshotBytes(t, e.states, c, before[c.String()])
	}
	for _, forbidden := range []string{"/Users/", "pkg/mod/", "goroutine ", "nativeCallContract"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("failure exposed %q: %v", forbidden, err)
		}
	}
}
