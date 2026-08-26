package runtime

import (
	"bytes"
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/state"
)

func TestUint256CanonicalToHexBoundaries(t *testing.T) {
	for _, tc := range []struct{ input, hex, code string }{{"0", "0x0", ""}, {"1", "0x1", ""}, {u256Max, "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", ""}, {"01", "", Uint256CanonicalDecimalError}, {u256Modulus, "", Uint256CanonicalDecimalError}} {
		hex, code := uint256CanonicalToHex(tc.input)
		if hex != tc.hex || code != tc.code {
			t.Fatalf("canonicalToHex(%q)=(%q,%q), want (%q,%q)", tc.input, hex, code, tc.hex, tc.code)
		}
	}
}

const uint256OptimizedContractSource = `package contract
import ("mitum/math/v1/u256";"mitum/chain")
var stored string
func Initialize(ctx chain.WriteContext) error { stored="0";return nil }
func Convert(ctx chain.WriteContext,v string) error { x,err:=u256.FromCanonicalDecimal(v);if err!=nil{return err};stored=u256.ToCanonicalDecimal(x);return nil }
func FromOnly(ctx chain.WriteContext,v string) error { _,err:=u256.FromCanonicalDecimal(v);return err }
func CheckOnly(ctx chain.WriteContext,v string) error { if u256.IsCanonicalDecimal(v){stored="1"};return nil }
func MulDivDecimal(ctx chain.WriteContext,a string,b string,d string) error { v,err:=u256.MulDivCanonicalDecimal(a,b,d);if err!=nil{return err};stored=v;return nil }
func SqrtDecimal(ctx chain.WriteContext,x string) error { v,err:=u256.SqrtCanonicalDecimal(x);if err!=nil{return err};stored=v;return nil }
func Get(ctx chain.QueryContext) string { return stored }
func IsCanonical(ctx chain.QueryContext,v string) bool { return u256.IsCanonicalDecimal(v) }
func QueryMulDiv(ctx chain.QueryContext,a string,b string,d string) string { v,err:=u256.MulDivCanonicalDecimal(a,b,d);if err!=nil{return err.Error()};return v }
func QuerySqrt(ctx chain.QueryContext,x string) string { v,err:=u256.SqrtCanonicalDecimal(x);if err!=nil{return err.Error()};return v }
func QueryFrom(ctx chain.QueryContext,v string) bool { _,err:=u256.FromCanonicalDecimal(v);return err==nil }
func QueryMulDivN(ctx chain.QueryContext,n uint64) string { out:="0";for i:=uint64(0);i<n;i++{v,err:=u256.MulDivCanonicalDecimal("6","7","4");if err!=nil{return err.Error()};out=v};return out }
func UintMulDiv(ctx chain.WriteContext,v string) error { x:=u256.MustFromCanonicalDecimal(v);z,err:=u256.MulDiv(x,x,x);if err!=nil{return err};stored=u256.ToCanonicalDecimal(z);return nil }
func UintSqrt(ctx chain.WriteContext,v string) error { x:=u256.MustFromCanonicalDecimal(v);stored=u256.ToCanonicalDecimal(u256.Sqrt(x));return nil }
`

func TestUint256OptimizedMaxWriteAndQuery(t *testing.T) {
	e := newOptimizedEnv(t)
	e.write(t, "Convert", map[string]string{"v": u256Max})
	if got := e.query(t, "Get", nil); got != u256Max {
		t.Fatalf("Convert max=%q", got)
	}
	q, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256OptimizedContractSource, Function: "IsCanonical", CallData: map[string]string{"v": u256Max}})
	if err != nil {
		t.Fatalf("max IsCanonical query: %v", err)
	}
	if got, ok := q.Result.(bool); !ok || !got {
		t.Fatalf("max IsCanonical=%#v", q.Result)
	}
	q, err = e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256OptimizedContractSource, Function: "QueryFrom", CallData: map[string]string{"v": u256Max}})
	if err != nil {
		t.Fatalf("max FromCanonical query: %v", err)
	}
	if got, ok := q.Result.(bool); !ok || !got {
		t.Fatalf("max QueryFrom=%#v", q.Result)
	}
	e.write(t, "MulDivDecimal", map[string]string{"a": u256Max, "b": u256Max, "d": u256Max})
	if got := e.query(t, "Get", nil); got != u256Max {
		t.Fatalf("MulDiv max=%q", got)
	}
	e.write(t, "SqrtDecimal", map[string]string{"x": u256Max})
	want := "340282366920938463463374607431768211455"
	if got := e.query(t, "Get", nil); got != want {
		t.Fatalf("Sqrt max=%q", got)
	}
	for _, tc := range []struct {
		fn   string
		data map[string]string
		want string
	}{{"QueryMulDiv", map[string]string{"a": u256Max, "b": u256Max, "d": u256Max}, u256Max}, {"QuerySqrt", map[string]string{"x": u256Max}, want}} {
		if got := e.query(t, tc.fn, tc.data); got != tc.want {
			t.Fatalf("%s=%q", tc.fn, got)
		}
	}
}

func TestUint256OptimizedUintMaximumPaths(t *testing.T) {
	e := newOptimizedEnv(t)
	before := e.snapshot(t)
	if _, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: e.height() + 1, ContractCode: uint256OptimizedContractSource, Function: "UintMulDiv", CallData: map[string]string{"v": u256Max}}); err == nil || !strings.Contains(err.Error(), "out of gas") {
		t.Fatalf("maximum UintMulDiv error=%v", err)
	}
	if !bytes.Equal(before, e.snapshot(t)) {
		t.Fatal("maximum UintMulDiv OOG changed snapshot")
	}
	e.write(t, "UintSqrt", map[string]string{"v": u256Max})
	if got, want := e.query(t, "Get", nil), "340282366920938463463374607431768211455"; got != want {
		t.Fatalf("maximum UintSqrt=%q, want %q", got, want)
	}
}

func TestUint256OptimizedErrorsRollBack(t *testing.T) {
	e := newOptimizedEnv(t)
	e.write(t, "Convert", map[string]string{"v": "7"})
	for _, tc := range []struct {
		fn     string
		data   map[string]string
		reason string
	}{{"Convert", map[string]string{"v": "01"}, Uint256CanonicalDecimalError}, {"MulDivDecimal", map[string]string{"a": "1", "b": "1", "d": "0"}, Uint256MulDivDenominatorZero}, {"MulDivDecimal", map[string]string{"a": u256Max, "b": u256Max, "d": "1"}, Uint256MulDivResultOverflow}} {
		before := e.snapshot(t)
		_, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: e.height() + 1, ContractCode: uint256OptimizedContractSource, Function: tc.fn, CallData: tc.data})
		if err == nil || !strings.Contains(err.Error(), tc.reason) {
			t.Fatalf("%s error=%v", tc.fn, err)
		}
		if !bytes.Equal(before, e.snapshot(t)) {
			t.Fatalf("%s changed snapshot", tc.fn)
		}
	}
}

func TestUint256OptimizedQueryOutOfGasIsReadOnly(t *testing.T) {
	e := newOptimizedEnv(t)
	before := e.snapshot(t)
	_, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256OptimizedContractSource, Function: "QueryMulDivN", CallData: map[string]string{"n": "100"}})
	if err == nil || !strings.Contains(err.Error(), "out of gas") {
		t.Fatalf("query OOG error=%v", err)
	}
	if !bytes.Equal(before, e.snapshot(t)) {
		t.Fatal("query OOG changed snapshot")
	}
	for _, forbidden := range []string{"/Users/", "pkg/mod/", "goroutine ", "gas descriptor"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("query OOG exposes %q: %v", forbidden, err)
		}
	}
}

type optimizedEnv struct {
	engine           ContractEngine
	states           map[string]base.State
	contract, sender base.Address
}

func newOptimizedEnv(t *testing.T) *optimizedEnv {
	t.Helper()
	e := &optimizedEnv{engine: NewGnoEngine(), states: map[string]base.State{}, contract: base.NewStringAddress("contractu256optimized"), sender: base.NewStringAddress("senderu256optimized")}
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeRegister, Contract: e.contract, Sender: e.sender, Height: 1, ContractCode: uint256OptimizedContractSource})
	if err != nil {
		t.Fatal(err)
	}
	applyStateMerges(e.states, 1, r.StateMerges)
	return e
}
func (e *optimizedEnv) height() base.Height {
	return e.states[state.SnapshotStateKey(e.contract)].Height()
}
func (e *optimizedEnv) write(t *testing.T, fn string, data map[string]string) {
	t.Helper()
	h := e.height() + 1
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: h, ContractCode: uint256OptimizedContractSource, Function: fn, CallData: data})
	if err != nil {
		t.Fatalf("%s: %v", fn, err)
	}
	applyStateMerges(e.states, h, r.StateMerges)
}
func (e *optimizedEnv) query(t *testing.T, fn string, data map[string]string) string {
	t.Helper()
	q, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256OptimizedContractSource, Function: fn, CallData: data})
	if err != nil {
		t.Fatalf("query %s: %v", fn, err)
	}
	v, ok := q.Result.(string)
	if !ok {
		t.Fatalf("query %s=%#v", fn, q.Result)
	}
	return v
}
func (e *optimizedEnv) snapshot(t *testing.T) []byte {
	t.Helper()
	v, err := state.GetSnapshotFromState(e.states[state.SnapshotStateKey(e.contract)])
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte(nil), v.Snapshot...)
}
