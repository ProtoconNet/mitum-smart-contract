package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
)

func TestClassifyGnoExecutionPanicOutOfGasSanitized(t *testing.T) {
	gasMeter := NewGnoGasMeter(1)
	gasMeter.ConsumeGas(1, "sensitive-gas-descriptor")

	err := ClassifyGnoExecutionPanic("gno execution", "raw-sensitive-gas-panic", gasMeter)
	assertPanicSurface(t, err.Error(), "gno execution out of gas", "raw-sensitive-gas-panic", "sensitive-gas-descriptor")
}

func TestClassifyGnoExecutionPanicAllocationSanitized(t *testing.T) {
	err := ClassifyGnoExecutionPanic("gno query", "allocator leaked raw-sensitive-allocation-detail", nil)
	assertPanicSurface(t, err.Error(), "gno query exceeded allocation limit", "raw-sensitive-allocation-detail", "allocator leaked")
}

func TestClassifyGnoExecutionPanicGenericSanitized(t *testing.T) {
	err := ClassifyGnoExecutionPanic("gno execution", map[string]string{"secret": "raw-sensitive-generic-detail"}, nil)
	assertPanicSurface(t, err.Error(), "gno execution panicked", "raw-sensitive-generic-detail", "secret")
}

func TestExecuteContractRecoverSanitizesPanicPayload(t *testing.T) {
	original := newGnoMachineAndPackageFunc
	newGnoMachineAndPackageFunc = func(
		*ExecutionContext,
		string,
		string,
		GnoExecutionLimits,
		gstore.GasMeter,
	) (*gno.Machine, *gno.PackageValue, error) {
		panic("raw-sensitive-execute-panic")
	}
	defer func() { newGnoMachineAndPackageFunc = original }()

	engine := NewGnoEngine()
	schema, err := engine.ValidateContract(negativeQueryContractSource)
	if err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}

	_, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     base.NewStringAddress("contractpanic001"),
		Sender:       base.NewStringAddress("senderpanic001"),
		Height:       base.Height(950),
		ContractCode: negativeQueryContractSource,
		Schema:       &schema,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected execute panic classification error")
	}
	assertPanicSurface(t, err.Error(), "gno execution panicked", "raw-sensitive-execute-panic")
}

func TestQueryContractRecoverSanitizesPanicPayload(t *testing.T) {
	original := newGnoMachineAndPackageFunc
	newGnoMachineAndPackageFunc = func(
		*ExecutionContext,
		string,
		string,
		GnoExecutionLimits,
		gstore.GasMeter,
	) (*gno.Machine, *gno.PackageValue, error) {
		panic("raw-sensitive-query-panic")
	}
	defer func() { newGnoMachineAndPackageFunc = original }()

	engine := NewGnoEngine()
	schema, err := engine.ValidateContract(negativeQueryContractSource)
	if err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}

	contract := base.NewStringAddress("contractpanic002")
	states := map[string]base.State{
		pstate.RuntimeStateKey(contract): common.NewBaseState(
			base.Height(951),
			pstate.RuntimeStateKey(contract),
			deriveRuntimeState(contract, negativeQueryContractSource),
			nil,
			nil,
		),
		pstate.SnapshotStateKey(contract): common.NewBaseState(
			base.Height(951),
			pstate.SnapshotStateKey(contract),
			pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, nil),
			nil,
			nil,
		),
	}

	_, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       base.NewStringAddress("senderpanic002"),
		Height:       base.Height(952),
		ContractCode: negativeQueryContractSource,
		Schema:       &schema,
		Function:     "GetValue",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected query panic classification error")
	}
	assertPanicSurface(t, err.Error(), "gno query panicked", "raw-sensitive-query-panic")
}

func assertPanicSurface(t *testing.T, got string, want string, forbidden ...string) {
	t.Helper()

	if !strings.Contains(got, want) {
		t.Fatalf("expected panic surface containing %q, got: %s", want, got)
	}
	for _, s := range forbidden {
		if strings.Contains(got, s) {
			t.Fatalf("panic surface leaked %q in: %s", s, got)
		}
	}
}
