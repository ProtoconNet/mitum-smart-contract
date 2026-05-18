package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
)

const hostABINativeGasContractSource = `package contract
import "mitum/chain"

var senderExists bool
var contractIsAccount bool

func Initialize(ctx chain.ContractContext) error {
	senderExists = chain.AccountExists(ctx.GetSender())
	contractIsAccount = chain.IsContractAccount(ctx.GetContract())
	return nil
}

func GetSenderExists(ctx chain.ContractContext) bool {
	return senderExists
}

func GetContractIsAccount(ctx chain.ContractContext) bool {
	return contractIsAccount
}

func DoesAccountExist(ctx chain.ContractContext, addr string) bool {
	return chain.AccountExists(addr)
}

func IsNamedContractAccount(ctx chain.ContractContext, addr string) bool {
	return chain.IsContractAccount(addr)
}
`

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
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
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
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
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
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
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
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
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

	senderAccount, err := types.NewAccount(sender, nil)
	if err != nil {
		t.Fatalf("NewAccount(sender) returned error: %v", err)
	}
	contractAccount, err := types.NewAccount(contract, nil)
	if err != nil {
		t.Fatalf("NewAccount(contract) returned error: %v", err)
	}

	return map[string]base.State{
		statecurrency.AccountStateKey(sender): common.NewBaseState(
			base.Height(1),
			statecurrency.AccountStateKey(sender),
			statecurrency.NewAccountStateValue(senderAccount),
			nil,
			nil,
		),
		statecurrency.AccountStateKey(contract): common.NewBaseState(
			base.Height(1),
			statecurrency.AccountStateKey(contract),
			statecurrency.NewAccountStateValue(contractAccount),
			nil,
			nil,
		),
		cestate.StateKeyContractAccount(contract): common.NewBaseState(
			base.Height(1),
			cestate.StateKeyContractAccount(contract),
			cestate.NewContractAccountStateValue(types.NewContractAccountStatus(sender, []base.Address{})),
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
