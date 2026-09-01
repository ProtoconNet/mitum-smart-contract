package runtime

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/state"
)

func TestUint256ContractExample(t *testing.T) {
	source := readUint256ContractExample(t)
	assertUint256ContractExamplePolicy(t, source)

	env := newUint256ExampleEnv(t, source, "100")
	if got := env.query(t, "GetTotal", nil).Result; got != "100" {
		t.Fatalf("GetTotal after Initialize = %#v", got)
	}

	env.write(t, "Add", map[string]string{"amount": "23"})
	if got := env.query(t, "GetTotal", nil).Result; got != "123" {
		t.Fatalf("GetTotal after Add = %#v", got)
	}

	env.write(t, "ApplyRatio", map[string]string{"numerator": "2", "denominator": "3"})
	if got := env.query(t, "GetTotal", nil).Result; got != "82" {
		t.Fatalf("GetTotal after ApplyRatio = %#v", got)
	}

	env.write(t, "StoreSqrt", nil)
	if got := env.query(t, "GetLastSqrt", nil).Result; got != "9" {
		t.Fatalf("GetLastSqrt = %#v", got)
	}
	if got := env.query(t, "PreviewSqrt", nil).Result; got != "9" {
		t.Fatalf("PreviewSqrt = %#v", got)
	}

	preview := env.query(t, "PreviewRatio", map[string]string{"numerator": "5", "denominator": "2"})
	if preview.Result != "205" || preview.Ok == nil || !*preview.Ok {
		t.Fatalf("PreviewRatio = %#v, ok=%v", preview.Result, preview.Ok)
	}
	assertCanonicalExampleResult(t, preview.Result)

	for _, tc := range []struct {
		function string
		callData map[string]string
	}{
		{"Add", map[string]string{"amount": "001"}},
		{"Reset", map[string]string{"next": "-1"}},
		{"ApplyRatio", map[string]string{"numerator": "1", "denominator": "0"}},
	} {
		before := env.snapshot(t)
		if err := env.writeError(t, tc.function, tc.callData); err == nil {
			t.Fatalf("%s unexpectedly succeeded", tc.function)
		}
		if !bytes.Equal(before, env.snapshot(t)) {
			t.Fatalf("%s changed state after failure", tc.function)
		}
	}

	failedPreview := env.query(t, "PreviewRatio", map[string]string{"numerator": "1", "denominator": "0"})
	if failedPreview.Result != "" || failedPreview.Ok == nil || *failedPreview.Ok {
		t.Fatalf("failed PreviewRatio = %#v, ok=%v", failedPreview.Result, failedPreview.Ok)
	}
	if got := env.query(t, "GetTotal", nil).Result; got != "82" {
		t.Fatalf("failed operations changed total to %#v", got)
	}
}

func readUint256ContractExample(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "example", "sc_uint256.go"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func assertUint256ContractExamplePolicy(t *testing.T, source string) {
	t.Helper()
	for _, required := range []string{
		`"mitum/chain"`,
		`"mitum/math/v1/u256"`,
		"var total string",
		"var lastSqrt string",
		"u256.FromCanonicalDecimal",
		"u256.ToCanonicalDecimal",
		"u256.MulDivCanonicalDecimal",
		"u256.SqrtCanonicalDecimal",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("example is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		`"gno.land/p/onbloc/uint256"`,
		"var total u256.Uint",
		"var total *u256.Uint",
		") u256.Uint",
		") *u256.Uint",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("example contains forbidden boundary %q", forbidden)
		}
	}
}

func assertCanonicalExampleResult(t *testing.T, result interface{}) {
	t.Helper()
	v, ok := result.(string)
	if !ok || v == "" || (len(v) > 1 && v[0] == '0') {
		t.Fatalf("result is not a canonical decimal string: %#v", result)
	}
}

type uint256ExampleEnv struct {
	engine           ContractEngine
	states           map[string]base.State
	contract, sender base.Address
	source           string
}

func newUint256ExampleEnv(t *testing.T, source, initial string) *uint256ExampleEnv {
	t.Helper()
	e := &uint256ExampleEnv{
		engine:   NewGnoEngine(),
		states:   map[string]base.State{},
		contract: base.NewStringAddress("contractu256example"),
		sender:   base.NewStringAddress("senderu256example"),
		source:   source,
	}
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{
		Mode: InvocationModeRegister, Contract: e.contract, Sender: e.sender, Height: 1,
		ContractCode: source, Function: "Initialize", CallData: map[string]string{"initial": initial},
	})
	if err != nil {
		t.Fatalf("register example: %v", err)
	}
	applyStateMerges(e.states, 1, r.StateMerges)
	return e
}

func (e *uint256ExampleEnv) height() base.Height {
	return e.states[state.SnapshotStateKey(e.contract)].Height()
}

func (e *uint256ExampleEnv) write(t *testing.T, function string, callData map[string]string) {
	t.Helper()
	if err := e.writeError(t, function, callData); err != nil {
		t.Fatalf("%s: %v", function, err)
	}
}

func (e *uint256ExampleEnv) writeError(t *testing.T, function string, callData map[string]string) error {
	t.Helper()
	height := e.height() + 1
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{
		Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: height,
		ContractCode: e.source, Function: function, CallData: callData,
	})
	if err == nil {
		applyStateMerges(e.states, height, r.StateMerges)
	}
	return err
}

func (e *uint256ExampleEnv) query(t *testing.T, function string, callData map[string]string) QueryResult {
	t.Helper()
	r, err := e.engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(e.states), QueryRequest{
		Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: e.source,
		Function: function, CallData: callData,
	})
	if err != nil {
		t.Fatalf("query %s: %v", function, err)
	}
	return r
}

func (e *uint256ExampleEnv) snapshot(t *testing.T) []byte {
	t.Helper()
	v, err := state.GetSnapshotFromState(e.states[state.SnapshotStateKey(e.contract)])
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte(nil), v.Snapshot...)
}
