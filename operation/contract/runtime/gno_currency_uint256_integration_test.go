package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	cstate "github.com/imfact-labs/currency-model/state/currency"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
)

const currencyUint256IntegrationSource = `package contract
import ("mitum/chain";"mitum/math/v1/u256")
var stored string
var observedBalance string
var observedAccount bool
var observedSender string
var observedCaller string
func Initialize(ctx chain.WriteContext) error { stored="initial";observedBalance="";return nil }
func UintDepositAndInspect(ctx chain.WriteContext,target string,currency string) error {
	amount,err:=u256.MulDivCanonicalDecimal("6","7","2");if err!=nil{return err}
	stored=amount;observedSender=ctx.GetSender();observedCaller=ctx.GetCaller()
	if err=chain.TransferSenderToContract(ctx,currency,amount);err!=nil{return err}
	return chain.CallContract(ctx,target,"Inspect",map[string]string{"addr":ctx.GetContract(),"currency":currency})
}
func Inspect(ctx chain.WriteContext,addr string,currency string) error { observedSender=ctx.GetSender();observedCaller=ctx.GetCaller();observedAccount=chain.AccountExists(addr);observedBalance,_=chain.BalanceOf(addr,currency);return nil }
func TryNestedTransfer(ctx chain.WriteContext,target string,currency string) error { return chain.CallContract(ctx,target,"NestedTransfer",map[string]string{"currency":currency}) }
func NestedTransfer(ctx chain.WriteContext,currency string) error { observedSender=ctx.GetSender();observedCaller=ctx.GetCaller();return chain.TransferSenderToContract(ctx,currency,"1") }
func CreateAndInspect(ctx chain.WriteContext,receiver string,target string,currency string) error { if err:=chain.TransferContractTo(ctx,receiver,currency,"40");err!=nil{return err};return chain.CallContract(ctx,target,"Inspect",map[string]string{"addr":receiver,"currency":currency}) }
func Set(ctx chain.WriteContext) error { stored="nested-changed";v,err:=u256.SqrtCanonicalDecimal("81");if err!=nil{return err};stored=v;return nil }
func Burn(ctx chain.WriteContext) error { stored="burning";for i:=0;i<200;i++{v,err:=u256.MulDivCanonicalDecimal("1","1","1");if err!=nil{return err};stored=v};return nil }
func AtomicOOG(ctx chain.WriteContext,receiver string,first string,second string,currency string) error { stored="caller-changed";if err:=chain.TransferSenderToContract(ctx,currency,"100");err!=nil{return err};if err:=chain.TransferContractTo(ctx,receiver,currency,"25");err!=nil{return err};if err:=chain.CallContract(ctx,first,"Set",map[string]string{});err!=nil{return err};return chain.CallContract(ctx,second,"Burn",map[string]string{}) }
func GetStored(ctx chain.QueryContext) string{return stored}
func GetBalance(ctx chain.QueryContext) string{return observedBalance}
func GetAccount(ctx chain.QueryContext) bool{return observedAccount}
func GetSender(ctx chain.QueryContext) string{return observedSender}
func GetCaller(ctx chain.QueryContext) string{return observedCaller}
`

type currencyUint256Env struct {
	states                        map[string]base.State
	caller, first, second, sender base.Address
}

func newCurrencyUint256Env(t *testing.T) *currencyUint256Env {
	t.Helper()
	e := &currencyUint256Env{states: map[string]base.State{}, caller: base.NewStringAddress("currencyu256call1"), first: base.NewStringAddress("currencyu256targ1"), second: base.NewStringAddress("currencyu256targ2"), sender: base.NewStringAddress("currencyu256send1")}
	registerNestedRuntimeContract(t, e.states, e.caller, e.sender, currencyUint256IntegrationSource, 1)
	registerNestedRuntimeContract(t, e.states, e.first, e.sender, currencyUint256IntegrationSource, 2)
	registerNestedRuntimeContract(t, e.states, e.second, e.sender, currencyUint256IntegrationSource, 3)
	addCurrencyState(t, e.states, "MCC", e.sender)
	addAccountState(t, e.states, e.sender)
	addContractCurrencyAccountState(t, e.states, e.caller, e.sender)
	addContractCurrencyAccountState(t, e.states, e.first, e.sender)
	addContractCurrencyAccountState(t, e.states, e.second, e.sender)
	addBalanceState(t, e.states, e.sender, "MCC", 1000)
	addBalanceState(t, e.states, e.caller, "MCC", 200)
	return e
}

func (e *currencyUint256Env) call(t *testing.T, fn string, data map[string]string) (ExecuteResult, base.OperationProcessReasonError) {
	t.Helper()
	return executeNestedCall(t, e.states, e.caller, e.sender, fn, data)
}
func queryCurrencyString(t *testing.T, e *currencyUint256Env, c base.Address, fn string) string {
	t.Helper()
	return mustQueryContract(t, e.states, c, e.sender, currencyUint256IntegrationSource, fn, nil).Result.(string)
}

func TestCurrencyUint256CalculatedTransfer(t *testing.T) {
	e := newCurrencyUint256Env(t)
	var r ExecuteResult
	var err base.OperationProcessReasonError
	meters, _ := recordNestedUint256Machines(t, func() {
		r, err = e.call(t, "UintDepositAndInspect", map[string]string{"target": e.first.String(), "currency": "MCC"})
	})
	if err != nil {
		t.Fatal(err)
	}
	applyStateMergesWithMerger(t, e.states, 100, r.StateMerges)
	assertRuntimeBalance(t, e.states, e.sender, "MCC", "979", true)
	assertRuntimeBalance(t, e.states, e.caller, "MCC", "221", true)
	if got := queryCurrencyString(t, e, e.caller, "GetStored"); got != "21" {
		t.Fatalf("amount=%q", got)
	}
	if got := queryCurrencyString(t, e, e.first, "GetBalance"); got != "221" {
		t.Fatalf("projected=%q", got)
	}
	if got := queryCurrencyString(t, e, e.caller, "GetSender"); got != e.sender.String() {
		t.Fatalf("sender=%q", got)
	}
	if got := queryCurrencyString(t, e, e.caller, "GetCaller"); got != e.sender.String() {
		t.Fatalf("caller=%q", got)
	}
	if int64(meters[0].GasConsumed()) <= mitumNativeCurrencyTransferGasBase+uint256NativeMulDivGasBase {
		t.Fatalf("combined gas=%d", meters[0].GasConsumed())
	}
}

func TestCurrencyCallerOriginRestriction(t *testing.T) {
	e := newCurrencyUint256Env(t)
	beforeCaller := snapshotBytesForContract(t, e.states, e.caller)
	beforeTarget := snapshotBytesForContract(t, e.states, e.first)
	r, err := e.call(t, "TryNestedTransfer", map[string]string{"target": e.first.String(), "currency": "MCC"})
	if err == nil || !strings.Contains(err.Error(), "allowed only from the top-level contract") {
		t.Fatalf("nested transfer=%v", err)
	}
	if len(r.StateMerges) != 0 {
		t.Fatalf("merges=%d", len(r.StateMerges))
	}
	assertSnapshotBytes(t, e.states, e.caller, beforeCaller)
	assertSnapshotBytes(t, e.states, e.first, beforeTarget)
	assertRuntimeBalance(t, e.states, e.sender, "MCC", "1000", true)
	assertRuntimeBalance(t, e.states, e.caller, "MCC", "200", true)
}

func TestCurrencyAccountOverlayVisibility(t *testing.T) {
	e := newCurrencyUint256Env(t)
	receiver := ctypes.NewAddress("0x2222222222222222222222222222222222222222")
	accountKey := cstate.AccountStateKey(receiver)
	balanceKey := cstate.BalanceStateKey(receiver, ctypes.CurrencyID("MCC"))
	r, err := e.call(t, "CreateAndInspect", map[string]string{"receiver": receiver.String(), "target": e.first.String(), "currency": "MCC"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.states[accountKey]; ok {
		t.Fatal("base account changed before merges")
	}
	if _, ok := e.states[balanceKey]; ok {
		t.Fatal("base balance changed before merges")
	}
	applyStateMergesWithMerger(t, e.states, 100, r.StateMerges)
	if got := mustQueryContract(t, e.states, e.first, e.sender, currencyUint256IntegrationSource, "GetAccount", nil).Result.(bool); !got {
		t.Fatal("nested AccountExists=false")
	}
	if got := queryCurrencyString(t, e, e.first, "GetBalance"); got != "40" {
		t.Fatalf("nested balance=%q", got)
	}
	if _, ok := e.states[accountKey]; !ok {
		t.Fatal("receiver account missing")
	}
	assertRuntimeBalance(t, e.states, receiver, "MCC", "40", true)
}

func TestCurrencyUint256AtomicOutOfGas(t *testing.T) {
	e := newCurrencyUint256Env(t)
	receiver := ctypes.NewAddress("0x3333333333333333333333333333333333333333")
	contracts := []base.Address{e.caller, e.first, e.second}
	before := map[string][]byte{}
	for _, c := range contracts {
		before[c.String()] = snapshotBytesForContract(t, e.states, c)
	}
	r, err := e.call(t, "AtomicOOG", map[string]string{"receiver": receiver.String(), "first": e.first.String(), "second": e.second.String(), "currency": "MCC"})
	if err == nil || !strings.Contains(err.Error(), "out of gas") {
		t.Fatalf("OOG=%v", err)
	}
	if len(r.StateMerges) != 0 {
		t.Fatalf("merges=%d", len(r.StateMerges))
	}
	for _, c := range contracts {
		assertSnapshotBytes(t, e.states, c, before[c.String()])
	}
	assertRuntimeBalance(t, e.states, e.sender, "MCC", "1000", true)
	assertRuntimeBalance(t, e.states, e.caller, "MCC", "200", true)
	if _, ok := e.states[cstate.AccountStateKey(receiver)]; ok {
		t.Fatal("receiver account persisted")
	}
	if _, ok := e.states[cstate.BalanceStateKey(receiver, ctypes.CurrencyID("MCC"))]; ok {
		t.Fatal("receiver balance persisted")
	}
	for _, s := range []string{"/Users/", "pkg/mod/", "gas descriptor", "nativeCallContract"} {
		if strings.Contains(err.Error(), s) {
			t.Fatalf("OOG exposed %q", s)
		}
	}
}

func TestCurrencyDeterministicMerges(t *testing.T) {
	run := func() ([]string, []string) {
		e := newCurrencyUint256Env(t)
		receiver := ctypes.NewAddress("0x4444444444444444444444444444444444444444")
		r, err := e.call(t, "CreateAndInspect", map[string]string{"receiver": receiver.String(), "target": e.first.String(), "currency": "MCC"})
		if err != nil {
			t.Fatal(err)
		}
		keys := make([]string, len(r.StateMerges))
		values := make([]string, len(r.StateMerges))
		for i, m := range r.StateMerges {
			keys[i] = m.Key()
			b, err := json.Marshal(m.Value())
			if err != nil {
				t.Fatal(err)
			}
			values[i] = string(b)
		}
		return keys, values
	}
	k1, v1 := run()
	k2, v2 := run()
	if fmt.Sprint(k1) != fmt.Sprint(k2) || fmt.Sprint(v1) != fmt.Sprint(v2) {
		t.Fatalf("merges differ\n%v\n%v\n%v\n%v", k1, k2, v1, v2)
	}
}
