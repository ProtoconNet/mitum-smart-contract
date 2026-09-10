package runtime

import (
	"bytes"
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/state"
)

const uint256CanonicalContractSource = `package contract
import ("mitum/math/v1/u256";"mitum/chain")
var totalSupply string
func Initialize(ctx chain.WriteContext) error { totalSupply=u256.ToCanonicalDecimal(u256.Zero());return nil }
func Set(ctx chain.WriteContext, input string) error { v,err:=u256.FromCanonicalDecimal(input);if err!=nil{return err};totalSupply=u256.ToCanonicalDecimal(v);return nil }
func Calculate(ctx chain.WriteContext, op string, a string, b string, d string) error {
	x,err:=u256.FromCanonicalDecimal(a);if err!=nil{return err}
	y,err:=u256.FromCanonicalDecimal(b);if err!=nil{return err}
	if op=="add" { totalSupply=u256.ToCanonicalDecimal(new(u256.Uint).Add(x,y));return nil }
	if op=="mul" { totalSupply=u256.ToCanonicalDecimal(new(u256.Uint).Mul(x,y));return nil }
	if op=="muldiv" { z,err:=u256.FromCanonicalDecimal(d);if err!=nil{return err};v,err:=u256.MulDiv(x,y,z);if err!=nil{return err};totalSupply=u256.ToCanonicalDecimal(v);return nil }
	totalSupply=u256.ToCanonicalDecimal(u256.Sqrt(x));return nil
}
func Get(ctx chain.QueryContext) string { return totalSupply }
func IsCanonical(ctx chain.QueryContext, input string) bool { return u256.IsCanonicalDecimal(input) }
`

func TestUint256CanonicalValidVectorsAndRoundTrip(t *testing.T) {
	env := newUint256CanonicalEnv(t, "contractu256canonical")
	for _, input := range []string{"0", "1", "42", u256Max} {
		env.write(t, "Set", map[string]string{"input": input})
		if got := env.queryString(t, "Get", nil); got != input {
			t.Fatalf("round trip %q = %q", input, got)
		}
		if input != u256Max && !env.queryBool(t, "IsCanonical", map[string]string{"input": input}) {
			t.Fatalf("canonical input rejected: %q", input)
		}
	}
}

func TestUint256CanonicalInvalidVectorsRollBack(t *testing.T) {
	env := newUint256CanonicalEnv(t, "contractu256invalid")
	env.write(t, "Set", map[string]string{"input": "7"})
	invalid := []string{"", "00", "01", "+0", "+1", "-0", "-1", " 1", "1 ", "1\n", "1.0", "1e3", "0x1", "١", "１２", u256Modulus, "115792089237316195423570985008687907853269984665640564039457584007913129639937", strings.Repeat("9", 256)}
	for _, input := range invalid {
		wantCode := Uint256InvalidDigitError
		switch {
		case input == "":
			wantCode = Uint256EmptyDecimalError
		case len(input) > 78:
			wantCode = Uint256DecimalTooLongError
		case len(input) > 1 && input[0] == '0':
			wantCode = Uint256LeadingZeroError
		case len(input) > 1 && input[0] == '+':
			wantCode = Uint256InvalidDigitError
		case input == u256Modulus || input == "115792089237316195423570985008687907853269984665640564039457584007913129639937":
			wantCode = Uint256DecimalOverflowError
		}
		before := env.snapshot(t)
		_, err := env.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(env.states), ExecuteRequest{
			Mode: InvocationModeCall, Contract: env.contract, Sender: env.sender, Height: env.height() + 1,
			ContractCode: uint256CanonicalContractSource, Function: "Set", CallData: map[string]string{"input": input},
		})
		if err == nil || !strings.Contains(err.Error(), wantCode) {
			t.Fatalf("Set(%q) error = %v, want %s", input, err, wantCode)
		}
		if !bytes.Equal(before, env.snapshot(t)) {
			t.Fatalf("Set(%q) changed snapshot", input)
		}
		if len(input) < 32 && env.queryBool(t, "IsCanonical", map[string]string{"input": input}) {
			t.Fatalf("IsCanonical accepted %q", input)
		}
	}
}

func TestUint256CanonicalArithmeticStorage(t *testing.T) {
	env := newUint256CanonicalEnv(t, "contractu256arithmetic")
	for _, tc := range []struct{ op, a, b, d, want string }{
		{"add", "40", "2", "1", "42"},
		{"mul", "6", "7", "1", "42"},
		{"muldiv", "6", "7", "4", "10"},
		{"sqrt", "81", "0", "1", "9"},
	} {
		env.write(t, "Calculate", map[string]string{"op": tc.op, "a": tc.a, "b": tc.b, "d": tc.d})
		if got := env.queryString(t, "Get", nil); got != tc.want {
			t.Fatalf("%s result = %q, want %q", tc.op, got, tc.want)
		}
	}
}

func TestUint256CanonicalSnapshotQueryAndDigestStability(t *testing.T) {
	direct := newUint256CanonicalEnv(t, "contractu256direct")
	direct.write(t, "Set", map[string]string{"input": "42"})
	directBytes := direct.snapshot(t)
	directValue := direct.snapshotValue(t)

	equivalent := newUint256CanonicalEnv(t, "contractu256equiv")
	equivalent.write(t, "Calculate", map[string]string{"op": "add", "a": "40", "b": "2", "d": "1"})
	equivalentBytes := equivalent.snapshot(t)
	equivalentValue := equivalent.snapshotValue(t)
	if !bytes.Equal(directBytes, equivalentBytes) {
		t.Fatalf("equivalent calculations produced different snapshots\ndirect: %s\nequivalent: %s", directBytes, equivalentBytes)
	}
	if !bytes.Equal(directValue.HashBytes(), equivalentValue.HashBytes()) {
		t.Fatal("equivalent snapshots produced different SnapshotStateValue hashes")
	}
	if sha256.Sum256(directBytes) != sha256.Sum256(equivalentBytes) {
		t.Fatal("equivalent snapshots produced different snapshot_sha256 values")
	}

	beforeQuery := direct.snapshot(t)
	if got := direct.queryString(t, "Get", nil); got != "42" {
		t.Fatalf("query result = %q", got)
	}
	if !bytes.Equal(beforeQuery, direct.snapshot(t)) {
		t.Fatal("query changed snapshot")
	}
	direct.write(t, "Set", map[string]string{"input": "42"})
	if !bytes.Equal(directBytes, direct.snapshot(t)) {
		t.Fatal("restore and recapture changed canonical snapshot bytes")
	}
}

func TestUint256CanonicalMustFailureIsSanitized(t *testing.T) {
	source := `package contract
import ("mitum/math/v1/u256";"mitum/chain")
func Initialize(ctx chain.WriteContext) error { _=u256.MustFromCanonicalDecimal("01");return nil }`
	_, err := NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{Mode: InvocationModeRegister, Contract: base.NewStringAddress("contractu256must"), Sender: base.NewStringAddress("senderu256must0001"), Height: 1, ContractCode: source})
	if err == nil {
		t.Fatal("expected MustFromCanonicalDecimal failure")
	}
	for _, forbidden := range []string{"/Users/", "pkg/mod/", "goroutine ", "runtime/debug"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("panic surface exposes %q: %v", forbidden, err)
		}
	}
}

type uint256CanonicalEnv struct {
	engine           ContractEngine
	states           map[string]base.State
	contract, sender base.Address
}

func newUint256CanonicalEnv(t *testing.T, address string) *uint256CanonicalEnv {
	t.Helper()
	e := &uint256CanonicalEnv{engine: NewGnoEngine(), states: map[string]base.State{}, contract: base.NewStringAddress(address), sender: base.NewStringAddress("senderu256canonical")}
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeRegister, Contract: e.contract, Sender: e.sender, Height: 1, ContractCode: uint256CanonicalContractSource})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	applyStateMerges(e.states, 1, r.StateMerges)
	return e
}
func (e *uint256CanonicalEnv) height() base.Height {
	return e.states[state.SnapshotStateKey(e.contract)].Height()
}
func (e *uint256CanonicalEnv) write(t *testing.T, function string, data map[string]string) {
	t.Helper()
	h := e.height() + 1
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: h, ContractCode: uint256CanonicalContractSource, Function: function, CallData: data})
	if err != nil {
		t.Fatalf("%s: %v", function, err)
	}
	applyStateMerges(e.states, h, r.StateMerges)
}
func (e *uint256CanonicalEnv) queryString(t *testing.T, function string, data map[string]string) string {
	t.Helper()
	q, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256CanonicalContractSource, Function: function, CallData: data})
	if err != nil {
		t.Fatalf("query %s: %v", function, err)
	}
	v, ok := q.Result.(string)
	if !ok {
		t.Fatalf("query %s result: %#v", function, q.Result)
	}
	return v
}
func (e *uint256CanonicalEnv) queryBool(t *testing.T, function string, data map[string]string) bool {
	t.Helper()
	q, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256CanonicalContractSource, Function: function, CallData: data})
	if err != nil {
		t.Fatalf("query %s: %v", function, err)
	}
	v, ok := q.Result.(bool)
	if !ok {
		t.Fatalf("query %s result: %#v", function, q.Result)
	}
	return v
}
func (e *uint256CanonicalEnv) snapshot(t *testing.T) []byte {
	t.Helper()
	return append([]byte(nil), e.snapshotValue(t).Snapshot...)
}
func (e *uint256CanonicalEnv) snapshotValue(t *testing.T) state.SnapshotStateValue {
	t.Helper()
	v, err := state.GetSnapshotFromState(e.states[state.SnapshotStateKey(e.contract)])
	if err != nil {
		t.Fatal(err)
	}
	return v
}
