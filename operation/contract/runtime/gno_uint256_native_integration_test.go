package runtime

import (
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/state"
)

const uint256NativeContractSource = `package contract
import ("mitum/math/v1/u256";"mitum/chain")
var stored string
func Initialize(ctx chain.WriteContext) error { stored="initial"; return nil }
func Run(ctx chain.WriteContext, op string, a string, b string, d string) error {
	if op=="muldiv" { v,err:=u256.MulDiv(u256.MustFromDecimal(a),u256.MustFromDecimal(b),u256.MustFromDecimal(d));if err!=nil{return err};stored=v.String();return nil }
	if op=="handled" { _,err:=u256.MulDiv(u256.MustFromDecimal(a),u256.MustFromDecimal(b),u256.MustFromDecimal(d));if err!=nil{stored=err.Error();return nil};return nil }
	if op=="sqrt" { stored=u256.Sqrt(u256.MustFromDecimal(a)).String();return nil }
	stored="must-roll-back";for i:=0;i<10000;i++{_,err:=u256.MulDiv(u256.One(),u256.One(),u256.One());if err!=nil{return err}};return nil
}
func Calculate(ctx chain.QueryContext, op string, a string, b string, d string) string {
	if op=="muldiv" { v,err:=u256.MulDiv(u256.MustFromDecimal(a),u256.MustFromDecimal(b),u256.MustFromDecimal(d));if err!=nil{return err.Error()};return v.String() }
	return u256.Sqrt(u256.MustFromDecimal(a)).String()
}
func Stored(ctx chain.QueryContext) string { return stored }
`

func TestUint256NativeWriteAndQueryIntegration(t *testing.T) {
	engine, states, contract, sender := newUint256NativeEnv(t)
	r, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode: InvocationModeCall, Contract: contract, Sender: sender, Height: 2,
		ContractCode: uint256NativeContractSource, Function: "Run",
		CallData: map[string]string{"op": "muldiv", "a": "6", "b": "7", "d": "4"},
	})
	if err != nil {
		t.Fatalf("write MulDiv: %v", err)
	}
	applyStateMerges(states, 2, r.StateMerges)

	query, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract: contract, Sender: sender, Height: 2, ContractCode: uint256NativeContractSource,
		Function: "Calculate", CallData: map[string]string{"op": "sqrt", "a": "81", "b": "0", "d": "1"},
	})
	if err != nil {
		t.Fatalf("query Sqrt: %v", err)
	}
	if got, ok := query.Result.(string); !ok || got != "9" {
		t.Fatalf("query Sqrt result: %#v", query.Result)
	}
}

func TestUint256NativeErrorsAndOutOfGasRollBack(t *testing.T) {
	engine, states, contract, sender := newUint256NativeEnv(t)
	r, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode: InvocationModeCall, Contract: contract, Sender: sender, Height: 2,
		ContractCode: uint256NativeContractSource, Function: "Run",
		CallData: map[string]string{"op": "handled", "a": "1", "b": "1", "d": "0"},
	})
	if err != nil {
		t.Fatalf("handled denominator zero: %v", err)
	}
	applyStateMerges(states, 2, r.StateMerges)

	for _, tc := range []struct{ op, a, b, d, reason string }{
		{"muldiv", "1", "1", "0", Uint256MulDivDenominatorZero},
		{"muldiv", "340282366920938463463374607431768211456", "340282366920938463463374607431768211456", "1", Uint256MulDivResultOverflow},
		{"oog", "0", "0", "1", "out of gas"},
	} {
		key := state.SnapshotStateKey(contract)
		height := states[key].Height()
		_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
			Mode: InvocationModeCall, Contract: contract, Sender: sender, Height: height + 1,
			ContractCode: uint256NativeContractSource, Function: "Run",
			CallData: map[string]string{"op": tc.op, "a": tc.a, "b": tc.b, "d": tc.d},
		})
		if err == nil || !strings.Contains(err.Error(), tc.reason) {
			t.Fatalf("%s error = %v, want reason %q", tc.op, err, tc.reason)
		}
		if got := states[key].Height(); got != height {
			t.Fatalf("%s changed snapshot height: %d -> %d", tc.op, height, got)
		}
		for _, forbidden := range []string{"/Users/", "pkg/mod/", "goroutine ", "runtime/debug"} {
			if strings.Contains(err.Error(), forbidden) {
				t.Fatalf("%s error exposes %q: %v", tc.op, forbidden, err)
			}
		}
	}
}

func TestUint256HiddenNativeBridgeCannotBeCalledByContract(t *testing.T) {
	source := `package contract
import ("mitum/math/v1/u256";"mitum/chain")
func Initialize(ctx chain.WriteContext) error { _,_ = u256._nativeMulDiv("1","1","1"); return nil }`
	engine := NewGnoEngine()
	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode: InvocationModeRegister, Contract: base.NewStringAddress("contractu256hidden"),
		Sender: base.NewStringAddress("senderu256hidden01"), Height: 1, ContractCode: source,
	})
	if err == nil {
		t.Fatal("expected hidden native bridge reference to fail package load")
	}
	if !strings.Contains(err.Error(), "_nativeMulDiv") {
		t.Fatalf("unexpected hidden bridge error: %v", err)
	}
}

func newUint256NativeEnv(t *testing.T) (ContractEngine, map[string]base.State, base.Address, base.Address) {
	t.Helper()
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractu256native")
	sender := base.NewStringAddress("senderu256native01")
	states := map[string]base.State{}
	r, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode: InvocationModeRegister, Contract: contract, Sender: sender, Height: 1, ContractCode: uint256NativeContractSource,
	})
	if err != nil {
		t.Fatalf("register native contract: %v", err)
	}
	applyStateMerges(states, 1, r.StateMerges)
	return engine, states, contract, sender
}
