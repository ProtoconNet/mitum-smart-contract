package runtime

import (
	"strings"
	"testing"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
)

const queryContextCurrentHeightSource = `package contract
import "mitum/chain"

var height int64

func Initialize(ctx chain.WriteContext) error {
	height = ctx.GetHeight()
	return nil
}

func StoreHeight(ctx chain.WriteContext) error {
	height = ctx.GetHeight()
	return nil
}

func GetWriteHeight(ctx chain.QueryContext) int64 {
	return height
}

func GetViewHeight(ctx chain.QueryContext) int64 {
	return ctx.GetHeight()
}

func GetCurrentHeight(ctx chain.QueryContext) int64 {
	return ctx.GetCurrentHeight()
}
`

const writeContextCurrentHeightMethodSource = `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error {
	_ = ctx.GetCurrentHeight()
	return nil
}
`

const chainCurrentHeightNativeSource = `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error {
	return nil
}

func GetCurrentHeight(ctx chain.QueryContext) int64 {
	return chain.CurrentHeight()
}
`

func TestQueryContextGetCurrentHeightAndViewHeight(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractcurh0001")
	sender := base.NewStringAddress("sendercurh00001")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(10),
		ContractCode: queryContextCurrentHeightSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(10), result.StateMerges)

	result, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(11),
		ContractCode: queryContextCurrentHeightSource,
		Function:     "StoreHeight",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(StoreHeight) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(11), result.StateMerges)

	viewHeight := states[pstate.SnapshotStateKey(contract)].Height()
	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        viewHeight,
		CurrentHeight: base.Height(99),
		ContractCode:  queryContextCurrentHeightSource,
		Function:      "GetViewHeight",
		CallData:      map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetViewHeight) returned error: %v", err)
	}
	if got, ok := qr.Result.(int64); !ok || got != int64(viewHeight) {
		t.Fatalf("expected query view height %d, got %#v", viewHeight, qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        viewHeight,
		CurrentHeight: base.Height(99),
		ContractCode:  queryContextCurrentHeightSource,
		Function:      "GetCurrentHeight",
		CallData:      map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetCurrentHeight) returned error: %v", err)
	}
	if got, ok := qr.Result.(int64); !ok || got != 99 {
		t.Fatalf("expected query current height 99, got %#v", qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       viewHeight,
		ContractCode: queryContextCurrentHeightSource,
		Function:     "GetWriteHeight",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetWriteHeight) returned error: %v", err)
	}
	if got, ok := qr.Result.(int64); !ok || got != 11 {
		t.Fatalf("expected write ctx.GetHeight result 11, got %#v", qr.Result)
	}
}

func TestWriteContextGetCurrentHeightRejected(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractcurh0002")
	sender := base.NewStringAddress("sendercurh00002")

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(10),
		ContractCode: writeContextCurrentHeightMethodSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	assertHeightABIRejected(t, err, "GetCurrentHeight")
}

func TestChainCurrentHeightNativeRemovedFromContractSurface(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractcurh0003")
	sender := base.NewStringAddress("sendercurh00003")

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(10),
		ContractCode: chainCurrentHeightNativeSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	assertHeightABIRejected(t, err, "CurrentHeight")
}

func assertHeightABIRejected(t *testing.T, err base.OperationProcessReasonError, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected %s ABI usage to fail", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error to mention %q, got %q", want, err.Error())
	}
}
