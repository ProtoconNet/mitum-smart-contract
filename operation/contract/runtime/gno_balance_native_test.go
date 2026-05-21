package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
)

const balanceNativeContractSource = `package contract
import "mitum/chain"

var initializedBalance string
var initializedBalanceOK bool
var storedBalance string
var storedBalanceOK bool

func Initialize(ctx chain.WriteContext, currency string) error {
	initializedBalance, initializedBalanceOK = chain.BalanceOf(ctx.GetSender(), currency)
	return nil
}

func StoreBalance(ctx chain.WriteContext, addr string, currency string) error {
	storedBalance, storedBalanceOK = chain.BalanceOf(addr, currency)
	return nil
}

func GetInitializedBalance(ctx chain.QueryContext) (string, bool) {
	return initializedBalance, initializedBalanceOK
}

func GetStoredBalance(ctx chain.QueryContext) (string, bool) {
	return storedBalance, storedBalanceOK
}

func LookupBalance(ctx chain.QueryContext, addr string, currency string) (string, bool) {
	return chain.BalanceOf(addr, currency)
}

func QueryHeight(ctx chain.QueryContext) int64 {
	return ctx.GetHeight()
}

func CurrentHeight(ctx chain.QueryContext) int64 {
	return ctx.GetCurrentHeight()
}
`

func TestBalanceOfNativeQuerySemantics(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractbal0001")
	sender := base.NewStringAddress("senderbal0001")
	missing := base.NewStringAddress("missingbal0001")
	noBalance := base.NewStringAddress("nobalance00001")
	zero := base.NewStringAddress("zerobalance001")
	states := balanceNativeStates(t, contract, sender, noBalance, zero)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: balanceNativeContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{"currency": "MCC"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	assertBalanceQuery(t, engine, states, contract, sender, "GetInitializedBalance", nil, "12345", true)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     sender.String(),
		"currency": "MCC",
	}, "12345", true)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     zero.String(),
		"currency": "ZERO",
	}, "0", true)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     noBalance.String(),
		"currency": "MCC",
	}, "", false)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     sender.String(),
		"currency": "NOPE",
	}, "", false)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     missing.String(),
		"currency": "MCC",
	}, "", false)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     "not-an-address",
		"currency": "MCC",
	}, "", false)
	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     sender.String(),
		"currency": "bad",
	}, "", false)
}

func TestBalanceOfNativeWriteSemanticsAndGasRegistration(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractbal0002")
	sender := base.NewStringAddress("senderbal0002")
	noBalance := base.NewStringAddress("nobalance00002")
	zero := base.NewStringAddress("zerobalance002")
	states := balanceNativeStates(t, contract, sender, noBalance, zero)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: balanceNativeContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{"currency": "MCC"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	result, err = engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeCall,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(101),
		ContractCode: balanceNativeContractSource,
		Function:     "StoreBalance",
		CallData: map[string]string{
			"addr":     zero.String(),
			"currency": "ZERO",
		},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(StoreBalance) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(101), result.StateMerges)

	assertBalanceQuery(t, engine, states, contract, sender, "GetStoredBalance", nil, "0", true)
}

func TestBalanceOfNativeIndependentFromHeightSemantics(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractbal0003")
	sender := base.NewStringAddress("senderbal0003")
	noBalance := base.NewStringAddress("nobalance00003")
	zero := base.NewStringAddress("zerobalance003")
	states := balanceNativeStates(t, contract, sender, noBalance, zero)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: balanceNativeContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{"currency": "MCC"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        states[pstate.SnapshotStateKey(contract)].Height(),
		CurrentHeight: base.Height(999),
		ContractCode:  balanceNativeContractSource,
		Function:      "QueryHeight",
		CallData:      map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(QueryHeight) returned error: %v", err)
	}
	if got, ok := qr.Result.(int64); !ok || got != 100 {
		t.Fatalf("expected view height 100, got %#v", qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        states[pstate.SnapshotStateKey(contract)].Height(),
		CurrentHeight: base.Height(999),
		ContractCode:  balanceNativeContractSource,
		Function:      "CurrentHeight",
		CallData:      map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(CurrentHeight) returned error: %v", err)
	}
	if got, ok := qr.Result.(int64); !ok || got != 999 {
		t.Fatalf("expected current height 999, got %#v", qr.Result)
	}

	assertBalanceQuery(t, engine, states, contract, sender, "LookupBalance", map[string]string{
		"addr":     sender.String(),
		"currency": "MCC",
	}, "12345", true)
}

func TestBalanceOfNativeDecodeFailureIsSanitized(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractbal0004")
	sender := base.NewStringAddress("senderbal0004")
	noBalance := base.NewStringAddress("nobalance00004")
	zero := base.NewStringAddress("zerobalance004")
	states := balanceNativeStates(t, contract, sender, noBalance, zero)

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(100),
		ContractCode: balanceNativeContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{"currency": "MCC"},
	})
	if err != nil {
		t.Fatalf("ExecuteContract(Initialize) returned error: %v", err)
	}
	applyStateMerges(states, base.Height(100), result.StateMerges)

	states[statecurrency.BalanceStateKey(sender, types.CurrencyID("MCC"))] = common.NewBaseState(
		base.Height(101),
		statecurrency.BalanceStateKey(sender, types.CurrencyID("MCC")),
		statecurrency.NewAccountStateValue(mustTestAccount(t, sender)),
		nil,
		nil,
	)

	_, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: balanceNativeContractSource,
		Function:     "LookupBalance",
		CallData: map[string]string{
			"addr":     sender.String(),
			"currency": "MCC",
		},
	})
	if err == nil {
		t.Fatalf("expected QueryContract to fail on corrupted balance state")
	}
	if !strings.Contains(err.Error(), "gno query panicked") {
		t.Fatalf("expected sanitized panic surface, got %q", err.Error())
	}
	for _, raw := range []string{"BalanceOf native call failed", "balance state decode failed", statecurrency.BalanceStateKey(sender, types.CurrencyID("MCC"))} {
		if strings.Contains(err.Error(), raw) {
			t.Fatalf("expected raw internal detail %q to be hidden, got %q", raw, err.Error())
		}
	}
}

func assertBalanceQuery(
	t *testing.T,
	engine ContractEngine,
	states map[string]base.State,
	contract base.Address,
	sender base.Address,
	function string,
	callData map[string]string,
	expectedValue string,
	expectedOK bool,
) {
	t.Helper()

	if callData == nil {
		callData = map[string]string{}
	}

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:      contract,
		Sender:        sender,
		Height:        states[pstate.SnapshotStateKey(contract)].Height(),
		CurrentHeight: base.Height(999),
		ContractCode:  balanceNativeContractSource,
		Function:      function,
		CallData:      callData,
	})
	if err != nil {
		t.Fatalf("QueryContract(%s) returned error: %v", function, err)
	}
	if got, ok := qr.Result.(string); !ok || got != expectedValue {
		t.Fatalf("expected %s result %q, got %#v", function, expectedValue, qr.Result)
	}
	if qr.Ok == nil {
		t.Fatalf("expected %s optional bool result, got nil", function)
	}
	if *qr.Ok != expectedOK {
		t.Fatalf("expected %s ok=%v, got %v", function, expectedOK, *qr.Ok)
	}
}

func balanceNativeStates(t *testing.T, contract, sender, noBalance, zero base.Address) map[string]base.State {
	t.Helper()

	states := hostABINativeStates(t, contract, sender)
	addAccountState(t, states, noBalance)
	addAccountState(t, states, zero)
	addCurrencyState(t, states, "MCC", sender)
	addCurrencyState(t, states, "ZERO", sender)
	addBalanceState(t, states, sender, "MCC", 12345)
	addBalanceState(t, states, zero, "ZERO", 0)

	return states
}

func addAccountState(t *testing.T, states map[string]base.State, addr base.Address) {
	t.Helper()

	account := mustTestAccount(t, addr)
	states[statecurrency.AccountStateKey(addr)] = common.NewBaseState(
		base.Height(1),
		statecurrency.AccountStateKey(addr),
		statecurrency.NewAccountStateValue(account),
		nil,
		[]util.Hash{},
	)
}

func addCurrencyState(t *testing.T, states map[string]base.State, cid string, genesis base.Address) {
	t.Helper()

	currencyID := types.CurrencyID(cid)
	design := types.NewCurrencyDesign(
		common.ZeroBig,
		currencyID,
		common.NewBig(9),
		genesis,
		types.NewCurrencyPolicy(common.ZeroBig, types.NewNilFeeer()),
	)
	states[statecurrency.DesignStateKey(currencyID)] = common.NewBaseState(
		base.Height(1),
		statecurrency.DesignStateKey(currencyID),
		statecurrency.NewCurrencyDesignStateValue(design),
		nil,
		[]util.Hash{},
	)
}

func addBalanceState(t *testing.T, states map[string]base.State, addr base.Address, cid string, amount int64) {
	t.Helper()

	currencyID := types.CurrencyID(cid)
	states[statecurrency.BalanceStateKey(addr, currencyID)] = common.NewBaseState(
		base.Height(1),
		statecurrency.BalanceStateKey(addr, currencyID),
		statecurrency.NewBalanceStateValue(types.NewAmount(common.NewBig(amount), currencyID)),
		nil,
		[]util.Hash{},
	)
}

func mustTestAccount(t *testing.T, addr base.Address) types.Account {
	t.Helper()

	account, err := types.NewAccount(addr, nil)
	if err != nil {
		t.Fatalf("NewAccount(%v) returned error: %v", addr, err)
	}

	return account
}
