package runtime

import (
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/state"
)

const (
	u256Modulus = "115792089237316195423570985008687907853269984665640564039457584007913129639936"
	u256Max     = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
)

const uint256ArithmeticContractSource = `package contract
import (
	"gno.land/p/onbloc/uint256"
	"mitum/chain"
	"strconv"
)

var stored string

func Initialize(ctx chain.WriteContext) error { stored = "0"; return nil }
func Stored(ctx chain.QueryContext) string { return stored }

func Eval(ctx chain.WriteContext, op string, a string, b string, c string, n uint64) error {
	if op == "basic" {
		x := uint256.NewUint(7); y := uint256.Zero().Set(x); one := uint256.One()
		all := uint256.Zero().SetAllOne(); clone := x.Clone(); x.Clear(); y.SetOne()
		stored = y.String()+"|"+one.String()+"|"+all.String()+"|"+clone.String()+"|"+x.String()
		return nil
	}
	if op == "decimal" {
		x, err := uint256.FromDecimal(a); if err != nil { stored = "invalid"; return nil }
		y := uint256.Zero(); if err := y.SetFromDecimal(a); err != nil { stored = "invalid"; return nil }
		stored = x.Dec()+"|"+y.String()+"|"+uint256.MustFromDecimal(a).String(); return nil
	}
	if op == "decimal-single" { x,err:=uint256.FromDecimal(a);if err!=nil{stored="invalid"}else{stored=x.String()};return nil }
	if op == "decimal-valid" { _, err := uint256.FromDecimal(a); stored = strconv.FormatBool(err == nil); return nil }
	if op == "hex" {
		x, err := uint256.FromHex(a); if err != nil { stored = "invalid"; return nil }
		y := uint256.Zero(); if err := y.SetFromHex(a); err != nil { stored = "invalid"; return nil }
		stored = x.String()+"|"+y.String()+"|"+uint256.MustFromHex(a).String(); return nil
	}
	if op == "hex-valid" { _, err := uint256.FromHex(a); stored = strconv.FormatBool(err == nil); return nil }
	if op == "must-decimal" { stored = uint256.MustFromDecimal(a).String(); return nil }
	if op == "must-hex" { stored = uint256.MustFromHex(a).String(); return nil }
	if op == "bytes" {
		b32 := []byte{0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0}
		b33 := []byte{0xff,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78,0x9a,0xbc,0xde,0xf0}
		stored = new(uint256.Uint).SetBytes(b32).String()+"|"+new(uint256.Uint).SetBytes(b33).String(); return nil
	}
	x := uint256.MustFromDecimal(a); y := uint256.MustFromDecimal(b); z := new(uint256.Uint)
	if op == "alias-add" { stored = x.Add(x,y).String(); return nil }
	if op == "alias-sub" { stored = x.Sub(x,y).String(); return nil }
	if op == "alias-mul" { stored = x.Mul(x,y).String(); return nil }
	if op == "add" { stored = z.Add(x,y).String(); return nil }
	if op == "sub" { stored = z.Sub(x,y).String(); return nil }
	if op == "neg" { stored = z.Neg(x).String(); return nil }
	if op == "mul" { stored = z.Mul(x,y).String(); return nil }
	if op == "div" { stored = z.Div(x,y).String(); return nil }
	if op == "mod" { stored = z.Mod(x,y).String(); return nil }
	if op == "exp" { stored = z.Exp(x,y).String(); return nil }
	if op == "add-overflow" { _, over := z.AddOverflow(x,y); stored = z.String()+"|"+strconv.FormatBool(over); return nil }
	if op == "sub-overflow" { _, over := z.SubOverflow(x,y); stored = z.String()+"|"+strconv.FormatBool(over); return nil }
	if op == "mul-overflow" { _, over := z.MulOverflow(x,y); stored = z.String()+"|"+strconv.FormatBool(over); return nil }
	if op == "mulmod" { stored = z.MulMod(x,y,uint256.MustFromDecimal(c)).String(); return nil }
	if op == "divmod" { q,r := z.DivMod(x,y,uint256.MustFromDecimal(c)); stored = q.String()+"|"+r.String(); return nil }
	if op == "compare" {
		u := y.Uint64(); stored = strconv.Itoa(x.Cmp(y))+"|"+strconv.FormatBool(x.Eq(y))+"|"+strconv.FormatBool(x.Neq(y))+"|"+
			strconv.FormatBool(x.Lt(y))+"|"+strconv.FormatBool(x.Lte(y))+"|"+strconv.FormatBool(x.Gt(y))+"|"+strconv.FormatBool(x.Gte(y))+"|"+
			strconv.FormatBool(x.LtUint64(u))+"|"+strconv.FormatBool(x.GtUint64(u))+"|"+strconv.FormatBool(x.IsZero())+"|"+strconv.Itoa(x.Sign())+"|"+strconv.FormatBool(x.Sgt(y)); return nil
	}
	if op == "or" { stored = z.Or(x,y).String(); return nil }
	if op == "and" { stored = z.And(x,y).String(); return nil }
	if op == "not" { stored = z.Not(x).String(); return nil }
	if op == "andnot" { stored = z.AndNot(x,y).String(); return nil }
	if op == "xor" { stored = z.Xor(x,y).String(); return nil }
	if op == "lsh" { stored = z.Lsh(x,uint(n)).String(); return nil }
	if op == "rsh" { stored = z.Rsh(x,uint(n)).String(); return nil }
	if op == "srsh" { stored = z.SRsh(x,uint(n)).String(); return nil }
	if op == "uint64" { v,over := x.Uint64WithOverflow(); stored = strconv.FormatUint(x.Uint64(),10)+"|"+strconv.FormatUint(v,10)+"|"+strconv.FormatBool(over)+"|"+strconv.FormatBool(x.IsUint64()); return nil }
	if op == "length" { stored = strconv.Itoa(x.BitLen())+"|"+strconv.Itoa(x.ByteLen())+"|"+x.Byte(y).String(); return nil }
	if op == "codec" {
		text,err := x.MarshalText(); if err != nil { stored="text-error"; return nil }; p:=uint256.Zero(); if err:=p.UnmarshalText(text);err!=nil{stored="unmarshal-text-error";return nil}
		json,err:=x.MarshalJSON();if err!=nil{stored="json-error";return nil};q:=uint256.Zero();if err:=q.UnmarshalJSON(json);err!=nil{stored="unmarshal-json-error";return nil};s:=uint256.Zero();if err:=s.Scan(a);err!=nil{stored="scan-error";return nil}
		stored=string(text)+"|"+p.String()+"|"+string(json)+"|"+q.String()+"|"+s.String();return nil
	}
	if op == "differential" { stored=z.Add(x,y).String()+"|"+new(uint256.Uint).Sub(x,y).String()+"|"+new(uint256.Uint).Mul(x,y).String()+"|"+new(uint256.Uint).Div(x,y).String()+"|"+new(uint256.Uint).Mod(x,y).String();return nil }
	stored = "unknown"; return nil
}
`

func TestUint256CreationAliasingAndWriteSnapshotRoundTrip(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	env.eval(t, "basic", "0", "0", "0", 0, "1|1|"+u256Max+"|7|0")
	env.eval(t, "alias-add", "7", "9", "0", 0, "16")
	env.eval(t, "alias-sub", "7", "9", "0", 0, "115792089237316195423570985008687907853269984665640564039457584007913129639934")
	env.eval(t, "alias-mul", u256Max, u256Max, "0", 0, "1")
}

func TestUint256ArithmeticOverflowAndZeroSemantics(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	for _, tc := range []struct{ op, a, b, c, want string }{
		{"add", u256Max, "1", "0", "0"}, {"add-overflow", u256Max, "1", "0", "0|true"},
		{"sub", "0", "1", "0", u256Max}, {"sub-overflow", "0", "1", "0", u256Max + "|true"},
		{"neg", "1", "0", "0", u256Max}, {"mul", u256Max, u256Max, "0", "1"},
		{"mul-overflow", u256Max, "2", "0", "115792089237316195423570985008687907853269984665640564039457584007913129639934|true"},
		{"div", "31337", "0", "0", "0"}, {"mod", "31337", "0", "0", "0"}, {"mulmod", "7", "9", "0", "0"},
		{"mulmod", "7", "9", "10", "3"}, {"divmod", "31337", "3", "99", "10445|2"},
		{"divmod", "31337", "0", "99", "0|0"}, {"exp", "2", "256", "0", "0"}, {"exp", "0", "0", "0", "1"},
	} {
		env.eval(t, tc.op, tc.a, tc.b, tc.c, 0, tc.want)
	}
}

func TestUint256ComparisonAndSignedGreaterThan(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	env.eval(t, "compare", "0", "0", "0", 0, "0|true|false|false|true|false|true|false|false|true|0|false")
	env.eval(t, "compare", "9", "7", "0", 0, "1|false|true|false|false|true|true|false|true|false|1|true")
	env.eval(t, "compare", u256Max, "0", "0", 0, "1|false|true|false|false|true|true|false|true|false|-1|false")
}

func TestUint256BitwiseAndShiftBoundaries(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	for _, tc := range []struct{ op, a, b, w string }{{"or", "10", "12", "14"}, {"and", "10", "12", "8"}, {"andnot", "10", "12", "2"}, {"xor", "10", "12", "6"}, {"not", "0", "0", u256Max}} {
		env.eval(t, tc.op, tc.a, tc.b, "0", 0, tc.w)
	}
	modulus := mustBigInt(u256Modulus)
	for _, n := range []uint64{0, 1, 63, 64, 127, 128, 255, 256, 257} {
		l := new(big.Int).Lsh(big.NewInt(1), uint(n))
		l.Mod(l, modulus)
		env.eval(t, "lsh", "1", "0", "0", n, l.String())
		r := new(big.Int).Rsh(mustBigInt(u256Max), uint(n))
		env.eval(t, "rsh", u256Max, "0", "0", n, r.String())
		env.eval(t, "srsh", u256Max, "0", "0", n, u256Max)
	}
}

func TestUint256ConversionBoundariesAndCodecs(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	for _, v := range []string{"0", "1", "18446744073709551615"} {
		env.eval(t, "decimal", v, "0", "0", 0, v+"|"+v+"|"+v)
	}
	for _, v := range []string{"18446744073709551616", "340282366920938463463374607431768211455", "340282366920938463463374607431768211456", u256Max} {
		env.eval(t, "decimal-single", v, "0", "0", 0, v)
	}
	for _, v := range []string{"", "-1", "invalid", u256Modulus} {
		env.eval(t, "decimal-valid", v, "0", "0", 0, "false")
	}
	env.eval(t, "hex", "0xffffffffffffffff", "0", "0", 0, "18446744073709551615|18446744073709551615|18446744073709551615")
	for _, v := range []string{"", "ff", "0x", "0x00", "0x10000000000000000000000000000000000000000000000000000000000000000"} {
		env.eval(t, "hex-valid", v, "0", "0", 0, "false")
	}
	env.eval(t, "uint64", "18446744073709551615", "0", "0", 0, "18446744073709551615|18446744073709551615|false|true")
	env.eval(t, "uint64", "18446744073709551616", "0", "0", 0, "0|0|true|false")
	env.eval(t, "length", "256", "30", "0", 0, "9|2|1")
	env.eval(t, "length", "256", "32", "0", 0, "9|2|0")
	bv := "8234104123542484900769178205574010627627573691361805720124810878238590820080"
	env.eval(t, "bytes", "0", "0", "0", 0, bv+"|"+bv)
	env.eval(t, "codec", "12345", "0", "0", 0, "12345|12345|\"12345\"|12345|12345")
}

func TestUint256DeterministicBigIntDifferential(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	rng := rand.New(rand.NewSource(256))
	modulus := mustBigInt(u256Modulus)
	vectorLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	for i := 0; i < 8; i++ {
		a := new(big.Int).Rand(rng, vectorLimit)
		b := new(big.Int).Rand(rng, vectorLimit)
		if b.Sign() == 0 {
			b.SetInt64(1)
		}
		add := new(big.Int).Mod(new(big.Int).Add(a, b), modulus)
		sub := new(big.Int).Mod(new(big.Int).Sub(a, b), modulus)
		mul := new(big.Int).Mod(new(big.Int).Mul(a, b), modulus)
		div := new(big.Int).Div(new(big.Int).Set(a), b)
		rem := new(big.Int).Mod(new(big.Int).Set(a), b)
		for _, tc := range []struct{ op, want string }{{"add", add.String()}, {"sub", sub.String()}, {"mul", mul.String()}, {"div", div.String()}, {"mod", rem.String()}} {
			env.eval(t, tc.op, a.String(), b.String(), "0", 0, tc.want)
		}
	}
}

func TestUint256ExternalTypesRemainOutsideSchemaABI(t *testing.T) {
	for _, decl := range []string{"var amount *uint256.Uint\nfunc Initialize(ctx chain.WriteContext) error{return nil}", "func Initialize(ctx chain.WriteContext, amount *uint256.Uint) error{return nil}", "func Initialize(ctx chain.WriteContext) error{return nil}\nfunc Amount(ctx chain.QueryContext) *uint256.Uint{return uint256.Zero()}"} {
		source := "package contract\nimport (\"gno.land/p/onbloc/uint256\";\"mitum/chain\")\n" + decl
		if _, err := AnalyzeContractSchema(source); err == nil {
			t.Fatalf("expected external Uint ABI rejection: %s", decl)
		}
	}
}

func TestUint256InvalidMustConversionsUseSanitizedRuntimeError(t *testing.T) {
	env := newUint256RuntimeEnv(t)
	for _, tc := range []struct{ op, input string }{{"must-decimal", u256Modulus}, {"must-hex", "0x10000000000000000000000000000000000000000000000000000000000000000"}} {
		height := env.states[state.SnapshotStateKey(env.contract)].Height()
		_, err := env.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(env.states), ExecuteRequest{Mode: InvocationModeCall, Contract: env.contract, Sender: env.sender, Height: height + 1, ContractCode: uint256ArithmeticContractSource, Function: "Eval", CallData: map[string]string{"op": tc.op, "a": tc.input, "b": "0", "c": "0", "n": "0"}})
		if err == nil {
			t.Fatalf("expected %s failure", tc.op)
		}
		for _, forbidden := range []string{"/Users/", "pkg/mod/", "goroutine ", "runtime/debug"} {
			if strings.Contains(err.Error(), forbidden) {
				t.Fatalf("%s error exposes %q: %v", tc.op, forbidden, err)
			}
		}
		if got := env.states[state.SnapshotStateKey(env.contract)].Height(); got != height {
			t.Fatalf("%s failure changed snapshot height: %d -> %d", tc.op, height, got)
		}
	}
}

func TestUint256WrapperTypeAliasProbe(t *testing.T) {
	source := `package contract
import ("gno.land/p/onbloc/uint256";"mitum/chain")
type Uint = uint256.Uint
func Initialize(ctx chain.WriteContext) error{return nil}
func Probe(ctx chain.QueryContext) string{return new(Uint).SetOne().String()}`
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractu256alias")
	sender := base.NewStringAddress("senderu256alias01")
	states := map[string]base.State{}
	r, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{Mode: InvocationModeRegister, Contract: contract, Sender: sender, Height: 1, ContractCode: source})
	if err != nil {
		t.Fatalf("type alias registration probe: %v", err)
	}
	applyStateMerges(states, 1, r.StateMerges)
	q, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{Contract: contract, Sender: sender, Height: 1, ContractCode: source, Function: "Probe", CallData: map[string]string{}})
	if err != nil {
		t.Fatalf("type alias query probe: %v", err)
	}
	if got, ok := q.Result.(string); !ok || got != "1" {
		t.Fatalf("type alias pointer method result: %#v", q.Result)
	}
}

type uint256RuntimeEnv struct {
	engine           ContractEngine
	states           map[string]base.State
	contract, sender base.Address
}

func newUint256RuntimeEnv(t *testing.T) *uint256RuntimeEnv {
	t.Helper()
	e := &uint256RuntimeEnv{engine: NewGnoEngine(), states: map[string]base.State{}, contract: base.NewStringAddress("contractuint256arith"), sender: base.NewStringAddress("senderuint256arith01")}
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeRegister, Contract: e.contract, Sender: e.sender, Height: 1, ContractCode: uint256ArithmeticContractSource})
	if err != nil {
		t.Fatalf("register uint256 contract: %v", err)
	}
	applyStateMerges(e.states, 1, r.StateMerges)
	return e
}

func (e *uint256RuntimeEnv) eval(t *testing.T, op, a, b, c string, n uint64, want string) {
	t.Helper()
	height := e.states[state.SnapshotStateKey(e.contract)].Height() + 1
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: height, ContractCode: uint256ArithmeticContractSource, Function: "Eval", CallData: map[string]string{"op": op, "a": a, "b": b, "c": c, "n": strconv.FormatUint(n, 10)}})
	if err != nil {
		t.Fatalf("Eval(%s): %v", op, err)
	}
	applyStateMerges(e.states, height, r.StateMerges)
	q, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{Contract: e.contract, Sender: e.sender, Height: height, ContractCode: uint256ArithmeticContractSource, Function: "Stored", CallData: map[string]string{}})
	if err != nil {
		t.Fatalf("Stored after %s: %v", op, err)
	}
	if got, ok := q.Result.(string); !ok || got != want {
		t.Fatalf("Eval(%s,%s,%s): got %#v, want %q", op, a, b, q.Result, want)
	}
}

func mustBigInt(value string) *big.Int {
	out, ok := new(big.Int).SetString(value, 10)
	if !ok {
		panic("invalid test integer")
	}
	return out
}
