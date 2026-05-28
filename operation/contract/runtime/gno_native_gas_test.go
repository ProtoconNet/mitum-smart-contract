package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const hostABINativeGasContractSource = `package contract
import "mitum/chain"

var senderExists bool
var contractIsAccount bool

func Initialize(ctx chain.WriteContext) error {
	senderExists = chain.AccountExists(ctx.GetSender())
	contractIsAccount = chain.IsContractAccount(ctx.GetContract())
	return nil
}

func GetSenderExists(ctx chain.QueryContext) bool {
	return senderExists
}

func GetContractIsAccount(ctx chain.QueryContext) bool {
	return contractIsAccount
}

func DoesAccountExist(ctx chain.QueryContext, addr string) bool {
	return chain.AccountExists(addr)
}

func IsNamedContractAccount(ctx chain.QueryContext, addr string) bool {
	return chain.IsContractAccount(addr)
}
`

func TestMitumNativeGasCalibrationTiers(t *testing.T) {
	expected := map[string]int64{
		"AccountExists":     3000,
		"IsContractAccount": 3000,
		"BalanceOf":         9000,
		"SHA3Sum256Base":    1000,
		"SHA3Sum256PerByte": 2,
	}

	got := map[string]int64{
		"AccountExists":     mitumNativeSingleLookupGasBase,
		"IsContractAccount": mitumNativeSingleLookupGasBase,
		"BalanceOf":         mitumNativeTripleLookupGasBase,
		"SHA3Sum256Base":    mitumNativeSHA3Sum256GasBase,
		"SHA3Sum256PerByte": mitumNativeSHA3Sum256GasPerByte,
	}

	for name, want := range expected {
		if got[name] != want {
			t.Fatalf("unexpected %s gas: got %d, want %d", name, got[name], want)
		}
	}

	accountExistsGas := got["AccountExists"]
	isContractAccountGas := got["IsContractAccount"]
	balanceOfGas := got["BalanceOf"]

	if !(balanceOfGas > accountExistsGas) {
		t.Fatalf("expected BalanceOf gas > AccountExists gas")
	}
	if !(balanceOfGas > isContractAccountGas) {
		t.Fatalf("expected BalanceOf gas > IsContractAccount gas")
	}
}

func TestMitumNativeSHA3GasScalesWithInputLength(t *testing.T) {
	if got := mitumNativeSHA3Sum256Gas(0); got != mitumNativeSHA3Sum256GasBase {
		t.Fatalf("expected SHA3 base gas for empty input, got %d", got)
	}
	if got, want := mitumNativeSHA3Sum256Gas(3), int64(1006); got != want {
		t.Fatalf("unexpected SHA3 gas for 3 bytes: got %d, want %d", got, want)
	}

	short := mitumNativeSHA3Sum256Gas(32)
	long := mitumNativeSHA3Sum256Gas(4096)
	if long <= short {
		t.Fatalf("expected SHA3 gas to increase with input length: short=%d long=%d", short, long)
	}
}

func TestMitumNativeGasRegistrationAllowsAccountExists(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractgas0001")
	sender := base.NewStringAddress("sendergas0001")
	states := hostABINativeStates(t, contract, sender)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(500),
		ContractCode: hostABINativeGasContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(500), result.StateMerges)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: hostABINativeGasContractSource,
		Function:     "GetSenderExists",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetSenderExists) returned error: %v", err)
	}
	if got, ok := qr.Result.(bool); !ok || !got {
		t.Fatalf("expected senderExists=true, got %#v", qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: hostABINativeGasContractSource,
		Function:     "DoesAccountExist",
		CallData: map[string]string{
			"addr": sender.String(),
		},
	})
	if err != nil {
		t.Fatalf("QueryContract(DoesAccountExist) returned error: %v", err)
	}
	if got, ok := qr.Result.(bool); !ok || !got {
		t.Fatalf("expected DoesAccountExist=true, got %#v", qr.Result)
	}
}

func TestMitumNativeGasRegistrationAllowsIsContractAccount(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractgas0002")
	sender := base.NewStringAddress("sendergas0002")
	states := hostABINativeStates(t, contract, sender)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(500),
		ContractCode: hostABINativeGasContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	applyStateMerges(states, base.Height(500), result.StateMerges)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: hostABINativeGasContractSource,
		Function:     "GetContractIsAccount",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetContractIsAccount) returned error: %v", err)
	}
	if got, ok := qr.Result.(bool); !ok || !got {
		t.Fatalf("expected contractIsAccount=true, got %#v", qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[state.SnapshotStateKey(contract)].Height(),
		ContractCode: hostABINativeGasContractSource,
		Function:     "IsNamedContractAccount",
		CallData: map[string]string{
			"addr": contract.String(),
		},
	})
	if err != nil {
		t.Fatalf("QueryContract(IsNamedContractAccount) returned error: %v", err)
	}
	if got, ok := qr.Result.(bool); !ok || !got {
		t.Fatalf("expected IsNamedContractAccount=true, got %#v", qr.Result)
	}
}

func hostABINativeStates(t *testing.T, contract, sender base.Address) map[string]base.State {
	t.Helper()

	senderAccount, err := ctypes.NewAccount(sender, nil)
	if err != nil {
		t.Fatalf("NewAccount(sender) returned error: %v", err)
	}
	contractAccount, err := ctypes.NewAccount(contract, nil)
	if err != nil {
		t.Fatalf("NewAccount(contract) returned error: %v", err)
	}

	return map[string]base.State{
		cstate.AccountStateKey(sender): common.NewBaseState(
			base.Height(1),
			cstate.AccountStateKey(sender),
			cstate.NewAccountStateValue(senderAccount),
			nil,
			nil,
		),
		cstate.AccountStateKey(contract): common.NewBaseState(
			base.Height(1),
			cstate.AccountStateKey(contract),
			cstate.NewAccountStateValue(contractAccount),
			nil,
			nil,
		),
		cestate.StateKeyContractAccount(contract): common.NewBaseState(
			base.Height(1),
			cestate.StateKeyContractAccount(contract),
			cestate.NewContractAccountStateValue(ctypes.NewContractAccountStatus(sender, []base.Address{})),
			nil,
			nil,
		),
	}
}

func applyStateMerges(states map[string]base.State, height base.Height, merges []base.StateMergeValue) {
	for _, merge := range merges {
		states[merge.Key()] = common.NewBaseState(
			height,
			merge.Key(),
			merge.Value(),
			nil,
			nil,
		)
	}
}
