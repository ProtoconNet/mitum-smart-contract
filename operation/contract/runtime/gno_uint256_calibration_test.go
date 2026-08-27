package runtime

import (
	"testing"

	"github.com/imfact-labs/smart-contract-model/state"
)

type uint256CalibrationSample struct{ gas, alloc int64 }

func TestUint256FinalCalibrationDeterminism(t *testing.T) {
	e := newOptimizedEnv(t)
	for _, tc := range []struct {
		name, fn string
		data     map[string]string
	}{{"canonical-max", "FromOnly", map[string]string{"v": u256Max}}, {"muldiv-max", "MulDivDecimal", map[string]string{"a": u256Max, "b": u256Max, "d": u256Max}}, {"sqrt-max", "SqrtDecimal", map[string]string{"x": u256Max}}} {
		first := measureUint256CalibrationWrite(t, e, tc.fn, tc.data, WriteGnoExecutionLimits())
		for i := 0; i < 2; i++ {
			if got := measureUint256CalibrationWrite(t, e, tc.fn, tc.data, WriteGnoExecutionLimits()); got != first {
				t.Fatalf("%s measurement changed: %#v != %#v", tc.name, got, first)
			}
		}
		t.Logf("UINT256_CALIBRATION write %-14s gas=%d alloc=%d", tc.name, first.gas, first.alloc)
	}
}

func TestUint256FinalCalibrationQueryAllocationStable(t *testing.T) {
	e := newOptimizedEnv(t)
	for _, tc := range []struct {
		name, fn string
		data     map[string]string
	}{{"muldiv-max", "QueryMulDiv", map[string]string{"a": u256Max, "b": u256Max, "d": u256Max}}, {"sqrt-max", "QuerySqrt", map[string]string{"x": u256Max}}} {
		first := measureUint256CalibrationQuery(t, e, tc.fn, tc.data)
		for i := 0; i < 2; i++ {
			if got := measureUint256CalibrationQuery(t, e, tc.fn, tc.data); got != first {
				t.Fatalf("%s query measurement changed: %#v != %#v", tc.name, got, first)
			}
		}
		if first.alloc >= GnoQueryMaxAllocBytes {
			t.Fatalf("%s query allocation %d exceeds %d", tc.name, first.alloc, GnoQueryMaxAllocBytes)
		}
		t.Logf("UINT256_CALIBRATION query %-14s gas=%d alloc=%d headroom=%d", tc.name, first.gas, first.alloc, GnoQueryMaxAllocBytes-first.alloc)
	}
}

func TestUint256CalibrationCustomAllocationLimitFails(t *testing.T) {
	e := newOptimizedEnv(t)
	before := e.snapshot(t)
	defer func() {
		if recover() == nil {
			t.Fatal("expected custom allocation limit failure")
		}
		if string(before) != string(e.snapshot(t)) {
			t.Fatal("custom allocation failure changed snapshot")
		}
	}()
	_ = measureUint256CalibrationWrite(t, e, "MulDivDecimal", map[string]string{"a": u256Max, "b": u256Max, "d": u256Max}, GnoExecutionLimits{GasLimit: GnoWriteGasLimit, MaxAllocBytes: 1_000_000})
}

func TestUint256CalibrationTwoOperationWrite(t *testing.T) {
	e := newOptimizedEnv(t)
	h := e.height() + 1
	r, err := e.engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(e.states), ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: h, ContractCode: uint256OptimizedContractSource, CallItems: []ExecuteCallItem{{Function: "MulDivDecimal", CallData: map[string]string{"a": u256Max, "b": u256Max, "d": u256Max}}, {Function: "SqrtDecimal", CallData: map[string]string{"x": u256Max}}}})
	if err != nil {
		t.Fatalf("two-operation write: %v", err)
	}
	applyStateMerges(e.states, h, r.StateMerges)
	if got, want := e.query(t, "Get", nil), "340282366920938463463374607431768211455"; got != want {
		t.Fatalf("two-operation result=%q, want %q", got, want)
	}
}

func BenchmarkUint256CanonicalToHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = uint256CanonicalToHex(u256Max)
	}
}
func BenchmarkUint256MulDivDecimal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = uint256MulDivDecimal(u256Max, u256Max, u256Max)
	}
}
func BenchmarkUint256SqrtDecimal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = uint256SqrtDecimal(u256Max)
	}
}
func BenchmarkUint256CanonicalParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parseCanonicalUint256(u256Max)
	}
}

func measureUint256CalibrationWrite(t *testing.T, e *optimizedEnv, fn string, data map[string]string, limits GnoExecutionLimits) uint256CalibrationSample {
	t.Helper()
	schema, err := AnalyzeContractSchema(uint256OptimizedContractSource)
	if err != nil {
		t.Fatal(err)
	}
	meter := NewGnoGasMeter(limits.GasLimit)
	ctx, err := NewExecutionContext(newRuntimeTestEncoders(t), stateGetter(e.states), e.contract, e.sender, e.height()+1, false)
	if err != nil {
		t.Fatal(err)
	}
	m, pkg, err := newGnoMachineAndPackage(ctx, deriveRuntimeState(e.contract, uint256OptimizedContractSource).PackagePath, uint256OptimizedContractSource, limits, meter)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := state.GetSnapshotFromState(e.states[state.SnapshotStateKey(e.contract)])
	if err != nil {
		t.Fatal(err)
	}
	if err = RestoreSnapshot(m, pkg, sv.Snapshot, schema); err != nil {
		t.Fatal(err)
	}
	if err = invokeTypedWrite(m, pkg, ExecuteRequest{Mode: InvocationModeCall, Contract: e.contract, Sender: e.sender, Height: e.height() + 1, ContractCode: uint256OptimizedContractSource, Function: fn, CallData: data}, schema); err != nil {
		t.Fatal(err)
	}
	if _, err = CaptureSnapshot(pkg, m.Store, schema); err != nil {
		t.Fatal(err)
	}
	_, alloc := m.Alloc.Status()
	return uint256CalibrationSample{int64(meter.GasConsumed()), alloc}
}

func measureUint256CalibrationQuery(t *testing.T, e *optimizedEnv, fn string, data map[string]string) uint256CalibrationSample {
	t.Helper()
	schema, err := AnalyzeContractSchema(uint256OptimizedContractSource)
	if err != nil {
		t.Fatal(err)
	}
	limits := QueryGnoExecutionLimits()
	meter := NewGnoGasMeter(limits.GasLimit)
	ctx, err := NewExecutionContext(newRuntimeTestEncoders(t), stateGetter(e.states), e.contract, e.sender, e.height(), true)
	if err != nil {
		t.Fatal(err)
	}
	m, pkg, err := newGnoMachineAndPackage(ctx, deriveRuntimeState(e.contract, uint256OptimizedContractSource).PackagePath, uint256OptimizedContractSource, limits, meter)
	if err != nil {
		t.Fatal(err)
	}
	sv, err := state.GetSnapshotFromState(e.states[state.SnapshotStateKey(e.contract)])
	if err != nil {
		t.Fatal(err)
	}
	if err = RestoreSnapshot(m, pkg, sv.Snapshot, schema); err != nil {
		t.Fatal(err)
	}
	if _, err = CaptureSnapshot(pkg, m.Store, schema); err != nil {
		t.Fatal(err)
	}
	if _, err = invokeTypedQuery(m, pkg, QueryRequest{Contract: e.contract, Sender: e.sender, Height: e.height(), ContractCode: uint256OptimizedContractSource, Function: fn, CallData: data}, schema); err != nil {
		t.Fatal(err)
	}
	if _, err = CaptureSnapshot(pkg, m.Store, schema); err != nil {
		t.Fatal(err)
	}
	_, alloc := m.Alloc.Status()
	return uint256CalibrationSample{int64(meter.GasConsumed()), alloc}
}
