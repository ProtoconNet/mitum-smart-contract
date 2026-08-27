package runtime

import (
	"strings"
	"testing"

	"github.com/imfact-labs/currency-model/common"
	cstate "github.com/imfact-labs/currency-model/state/currency"
	cestate "github.com/imfact-labs/currency-model/state/extension"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
	pstate "github.com/imfact-labs/smart-contract-model/state"
)

const currencyTransferContractSource = `package contract
import "mitum/chain"

var lastBalance string
var lastBalanceOK bool

func Initialize(ctx chain.WriteContext) error { return nil }

func Deposit(ctx chain.WriteContext, currency string, amount string) error {
	return chain.TransferSenderToContract(ctx, currency, amount)
}

func Withdraw(ctx chain.WriteContext, receiver string, currency string, amount string) error {
	return chain.TransferContractTo(ctx, receiver, currency, amount)
}

func DepositAndRecord(ctx chain.WriteContext, currency string, amount string) error {
	if err := chain.TransferSenderToContract(ctx, currency, amount); err != nil {
		return err
	}
	lastBalance, lastBalanceOK = chain.BalanceOf(ctx.GetContract(), currency)
	return nil
}

func DepositAndCallRecord(ctx chain.WriteContext, target string, currency string, amount string) error {
	if err := chain.TransferSenderToContract(ctx, currency, amount); err != nil {
		return err
	}
	return chain.CallContract(ctx, target, "RecordBalance", map[string]string{"addr": ctx.GetContract(), "currency": currency})
}

func RecordBalance(ctx chain.WriteContext, addr string, currency string) error {
	lastBalance, lastBalanceOK = chain.BalanceOf(addr, currency)
	return nil
}

func NestedDeposit(ctx chain.WriteContext, target string, currency string, amount string) error {
	return chain.CallContract(ctx, target, "Deposit", map[string]string{"currency": currency, "amount": amount})
}

func DepositThenFail(ctx chain.WriteContext, currency string, amount string) error {
	if err := chain.TransferSenderToContract(ctx, currency, amount); err != nil {
		return err
	}
	panic("raw transfer failure")
}

func GetLastBalance(ctx chain.QueryContext) (string, bool) {
	return lastBalance, lastBalanceOK
}
`

func TestCurrencyTransferSenderToContractUsesOverlay(t *testing.T) {
	states, contract, _, sender := prepareCurrencyTransferState(t)

	result, err := executeNestedCall(t, states, contract, sender, "DepositAndRecord", map[string]string{
		"currency": "MCC",
		"amount":   "125",
	})
	if err != nil {
		t.Fatalf("ExecuteContract(DepositAndRecord) returned error: %v", err)
	}
	applyStateMergesWithMerger(t, states, base.Height(100), result.StateMerges)

	assertRuntimeBalance(t, states, sender, "MCC", "875", true)
	assertRuntimeBalance(t, states, contract, "MCC", "325", true)
	assertLastRecordedBalance(t, states, contract, sender, "325", true)
}

func TestCurrencyTransferContractToAutoCreatesReceiver(t *testing.T) {
	states, contract, _, sender := prepareCurrencyTransferState(t)
	receiver := ctypes.NewAddress("0x1111111111111111111111111111111111111111")
	if err := receiver.IsValid(nil); err != nil {
		t.Fatalf("receiver test address is invalid: %v", err)
	}

	result, err := executeNestedCall(t, states, contract, sender, "Withdraw", map[string]string{
		"receiver": receiver.String(),
		"currency": "MCC",
		"amount":   "40",
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Withdraw) returned error: %v", err)
	}
	applyStateMergesWithMerger(t, states, base.Height(100), result.StateMerges)

	assertRuntimeBalance(t, states, contract, "MCC", "160", true)
	assertRuntimeBalance(t, states, receiver, "MCC", "40", true)
	if _, found := states[cstate.AccountStateKey(receiver)]; !found {
		t.Fatalf("expected receiver account to be auto-created")
	}
}

func TestCurrencyTransferNestedBalanceOfSeesProjectedBalance(t *testing.T) {
	states, contract, target, sender := prepareCurrencyTransferState(t)

	result, err := executeNestedCall(t, states, contract, sender, "DepositAndCallRecord", map[string]string{
		"target":   target.String(),
		"currency": "MCC",
		"amount":   "75",
	})
	if err != nil {
		t.Fatalf("ExecuteContract(DepositAndCallRecord) returned error: %v", err)
	}
	applyStateMergesWithMerger(t, states, base.Height(100), result.StateMerges)

	assertLastRecordedBalance(t, states, target, sender, "275", true)
	assertRuntimeBalance(t, states, contract, "MCC", "275", true)
}

func TestCurrencyTransferRejectsNestedRegisterAndQuery(t *testing.T) {
	states, contract, target, sender := prepareCurrencyTransferState(t)

	_, err := executeNestedCall(t, states, contract, sender, "NestedDeposit", map[string]string{
		"target":   target.String(),
		"currency": "MCC",
		"amount":   "1",
	})
	if err == nil || !strings.Contains(err.Error(), "allowed only from the top-level contract") {
		t.Fatalf("expected nested transfer rejection, got %v", err)
	}

	registerTransferSource := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error {
	return chain.TransferSenderToContract(ctx, "MCC", "1")
}
`
	registerContract := base.NewStringAddress("registertransfer1")
	_, err = NewGnoEngine().ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     registerContract,
		Sender:       sender,
		Height:       base.Height(101),
		ContractCode: registerTransferSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err == nil || !strings.Contains(err.Error(), "not allowed during contract registration") {
		t.Fatalf("expected register transfer rejection, got %v", err)
	}

	queryTransferSource := `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func QueryTransfer(ctx chain.QueryContext, currency string, amount string) string {
	_ = chain.TransferSenderToContract(ctx, currency, amount)
	return ""
}
`
	_, err = NewGnoEngine().QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: queryTransferSource,
		Function:     "QueryTransfer",
		CallData:     map[string]string{"currency": "MCC", "amount": "1"},
	})
	if err == nil {
		t.Fatalf("expected query transfer use to be rejected")
	}
}

func TestCurrencyTransferFailureRollsBackCurrencyMerges(t *testing.T) {
	states, contract, _, sender := prepareCurrencyTransferState(t)

	result, err := executeNestedCall(t, states, contract, sender, "DepositThenFail", map[string]string{
		"currency": "MCC",
		"amount":   "99",
	})
	if err == nil {
		t.Fatalf("expected transfer failure")
	}
	if len(result.StateMerges) != 0 {
		t.Fatalf("expected no state merges on failed transfer, got %d", len(result.StateMerges))
	}
	if strings.Contains(err.Error(), "raw transfer failure") {
		t.Fatalf("failure reason leaked raw panic payload: %v", err)
	}

	assertRuntimeBalance(t, states, sender, "MCC", "1000", true)
	assertRuntimeBalance(t, states, contract, "MCC", "200", true)
}

func prepareCurrencyTransferState(
	t *testing.T,
) (map[string]base.State, base.Address, base.Address, base.Address) {
	t.Helper()

	states, contract, target, sender := prepareNestedRuntimeState(t, currencyTransferContractSource)
	addCurrencyState(t, states, "MCC", sender)
	addAccountState(t, states, sender)
	addContractCurrencyAccountState(t, states, contract, sender)
	addContractCurrencyAccountState(t, states, target, sender)
	addBalanceState(t, states, sender, "MCC", 1000)
	addBalanceState(t, states, contract, "MCC", 200)

	return states, contract, target, sender
}

func addContractCurrencyAccountState(t *testing.T, states map[string]base.State, contract, owner base.Address) {
	t.Helper()

	addAccountState(t, states, contract)
	states[cestate.StateKeyContractAccount(contract)] = common.NewBaseState(
		base.Height(1),
		cestate.StateKeyContractAccount(contract),
		cestate.NewContractAccountStateValue(ctypes.NewContractAccountStatus(owner, []base.Address{})),
		nil,
		[]util.Hash{},
	)
}

func applyStateMergesWithMerger(
	t *testing.T,
	states map[string]base.State,
	height base.Height,
	merges []base.StateMergeValue,
) {
	t.Helper()

	for _, merge := range merges {
		merger := merge.Merger(height, states[merge.Key()])
		if err := merger.Merge(merge.Value(), nil); err != nil {
			t.Fatalf("Merge(%s) returned error: %v", merge.Key(), err)
		}
		st, err := merger.CloseValue()
		if err != nil {
			t.Fatalf("CloseValue(%s) returned error: %v", merge.Key(), err)
		}
		states[merge.Key()] = st
	}
}

func assertRuntimeBalance(
	t *testing.T,
	states map[string]base.State,
	address base.Address,
	currency string,
	expected string,
	expectedOK bool,
) {
	t.Helper()

	reader := StateBalanceReader{
		encs:         newRuntimeTestEncoders(t),
		getStateFunc: stateGetter(states),
	}
	got, ok, err := reader.BalanceOf(address.String(), currency)
	if err != nil {
		t.Fatalf("BalanceOf(%s, %s) returned error: %v", address, currency, err)
	}
	if got != expected || ok != expectedOK {
		t.Fatalf("unexpected balance for %s %s: got (%q, %v), want (%q, %v)", address, currency, got, ok, expected, expectedOK)
	}
}

func assertLastRecordedBalance(
	t *testing.T,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	expected string,
	expectedOK bool,
) {
	t.Helper()

	qr := mustQueryContract(t, states, contract, sender, currencyTransferContractSource, "GetLastBalance", map[string]string{})
	if got, ok := qr.Result.(string); !ok || got != expected {
		t.Fatalf("expected recorded balance %q, got %#v", expected, qr.Result)
	}
	if qr.Ok == nil || *qr.Ok != expectedOK {
		t.Fatalf("expected recorded balance ok=%v, got %v", expectedOK, qr.Ok)
	}
}
