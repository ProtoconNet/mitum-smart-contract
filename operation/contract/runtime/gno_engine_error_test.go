package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const typedWriteErrorContractSource = `package contract
import "mitum/chain"

type contractError string

func (e contractError) Error() string { return string(e) }

var owner string
var value string

func Initialize(ctx chain.ContractContext) error {
	owner = ctx.GetSender()
	return nil
}

func CreateData(ctx chain.ContractContext, data string) error {
	if ctx.GetSender() != owner {
		return contractError("only owner can create data")
	}
	if value != "" {
		return contractError("data already exists")
	}
	value = data
	return nil
}

func UpdateData(ctx chain.ContractContext, data string) error {
	if ctx.GetSender() != owner {
		return contractError("only owner can update data")
	}
	if value == "" {
		return contractError("data does not exist")
	}
	value = data
	return nil
}
`

func TestExecuteContractIncludesTypedWriteErrorMessage(t *testing.T) {
	engine := NewGnoEngine()
	owner := base.NewStringAddress("owner0001")
	contract := base.NewStringAddress("contracte0001")

	states := registerContractForWriteErrorTest(t, engine, typedWriteErrorContractSource, contract, owner)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       owner,
		Height:       base.Height(101),
		ContractCode: typedWriteErrorContractSource,
		Function:     "UpdateData",
		CallData: map[string]string{
			"data": "next",
		},
	})
	if err == nil {
		t.Fatalf("expected UpdateData error")
	}
	if got := err.Error(); !strings.Contains(got, `typed write function "UpdateData" returned error: data does not exist`) {
		t.Fatalf("expected contract error message in final error, got: %v", err)
	}
}

func TestExecuteContractIncludesTypedWriteOwnerErrorMessage(t *testing.T) {
	engine := NewGnoEngine()
	owner := base.NewStringAddress("owner0002")
	other := base.NewStringAddress("other0002")
	contract := base.NewStringAddress("contracte0002")

	states := registerContractForWriteErrorTest(t, engine, typedWriteErrorContractSource, contract, owner)

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       other,
		Height:       base.Height(101),
		ContractCode: typedWriteErrorContractSource,
		Function:     "UpdateData",
		CallData: map[string]string{
			"data": "next",
		},
	})
	if err == nil {
		t.Fatalf("expected UpdateData owner error")
	}
	if got := err.Error(); !strings.Contains(got, `typed write function "UpdateData" returned error: only owner can update data`) {
		t.Fatalf("expected owner error message in final error, got: %v", err)
	}
}

func TestExecuteContractTypedWriteNilErrorStillSucceeds(t *testing.T) {
	engine := NewGnoEngine()
	owner := base.NewStringAddress("owner0003")
	contract := base.NewStringAddress("contracte0004")

	states := registerContractForWriteErrorTest(t, engine, typedWriteErrorContractSource, contract, owner)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       owner,
		Height:       base.Height(101),
		ContractCode: typedWriteErrorContractSource,
		Function:     "CreateData",
		CallData: map[string]string{
			"data": "hello",
		},
	})
	if err != nil {
		t.Fatalf("expected CreateData success, got error: %v", err)
	}
	if len(result.StateMerges) == 0 {
		t.Fatalf("expected CreateData to produce state merges")
	}
}

func TestInvokeTypedWriteKeepsResultCountMismatchError(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	contract := base.NewStringAddress("contracte0003")
	sender := base.NewStringAddress("sendere0003")
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func Bad(ctx chain.ContractContext) {}
`

	execCtx, err := NewExecutionContext(encs, func(string) (base.State, bool, error) {
		return nil, false, nil
	}, contract, sender, base.Height(10), false)
	if err != nil {
		t.Fatalf("NewExecutionContext returned error: %v", err)
	}

	schema := ContractSchema{
		PackageName: "contract",
		Mode:        SchemaModeTypedArgs,
		Functions: []FunctionSchema{
			{
				Name:     "Bad",
				Exported: true,
				Params: []ParamSchema{
					{Name: "ctx", Type: TypeRef{Kind: TypeOpaque, Raw: "chain.ContractContext"}},
				},
			},
		},
	}

	limits := WriteGnoExecutionLimits()
	gasMeter := NewGnoGasMeter(limits.GasLimit)
	m, pkg, err := newGnoMachineAndPackage(execCtx, deriveRuntimeState(contract, source).PackagePath, source, limits, gasMeter)
	if err != nil {
		t.Fatalf("newGnoMachineAndPackage returned error: %v", err)
	}

	err = invokeTypedWrite(m, pkg, ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(10),
		ContractCode: source,
		Function:     "Bad",
		CallData:     map[string]string{},
	}, schema)
	if err == nil {
		t.Fatalf("expected result count mismatch error")
	}
	if got := err.Error(); !strings.Contains(got, `typed write function "Bad" must return exactly one error result`) {
		t.Fatalf("unexpected mismatch error: %v", err)
	}
}

func registerContractForWriteErrorTest(
	t *testing.T,
	engine ContractEngine,
	source string,
	contract base.Address,
	owner base.Address,
) map[string]base.State {
	t.Helper()

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       owner,
		Height:       base.Height(100),
		ContractCode: source,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("register ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	for _, merge := range result.StateMerges {
		states[merge.Key()] = common.NewBaseState(
			base.Height(100),
			merge.Key(),
			merge.Value(),
			nil,
			nil,
		)
	}

	if _, found := states[pstate.RuntimeStateKey(contract)]; !found {
		t.Fatalf("runtime state not created")
	}
	if _, found := states[pstate.SnapshotStateKey(contract)]; !found {
		t.Fatalf("snapshot state not created")
	}

	return states
}
