package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

func (_ GenesisCurrencies) PreProcess(
	ctx context.Context, _ mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	return ctx, nil, nil
}

func (op GenesisCurrencies) Process(
	_ context.Context, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(GenesisCurrenciesFact)
	if !ok {
		return nil, nil, errors.Errorf("expected GenesisCurrenciesFact, not %T", op.Fact())
	}

	newAddress, err := fact.Address()
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError(err.Error()), nil
	}

	ns, err := state.NotExistsState(currency.StateKeyAccount(newAddress), "key of genesis", getStateFunc)
	if err != nil {
		return nil, nil, err
	}

	cs := make([]base.CurrencyDesign, len(fact.cs))
	gas := map[base.CurrencyID]mitumbase.StateMergeValue{}
	sts := map[base.CurrencyID]mitumbase.StateMergeValue{}
	for i := range fact.cs {
		c := fact.cs[i]
		c.SetGenesisAccount(newAddress)
		cs[i] = c

		st, err := state.NotExistsState(currency.StateKeyCurrencyDesign(c.Currency()), "currency", getStateFunc)
		if err != nil {
			return nil, nil, err
		}

		sts[c.Currency()] = state.NewStateMergeValue(st.Key(), currency.NewCurrencyDesignStateValue(c))

		st, err = state.NotExistsState(currency.StateKeyBalance(newAddress, c.Currency()), "balance of genesis", getStateFunc)
		if err != nil {
			return nil, nil, err
		}
		gas[c.Currency()] = state.NewStateMergeValue(st.Key(), currency.NewBalanceStateValue(base.NewZeroAmount(c.Currency())))
	}

	var smvs []mitumbase.StateMergeValue
	if ac, err := base.NewAccount(newAddress, fact.keys); err != nil {
		return nil, nil, err
	} else {
		smvs = append(smvs, state.NewStateMergeValue(ns.Key(), currency.NewAccountStateValue(ac)))
	}

	for i := range cs {
		c := cs[i]
		v, ok := gas[c.Currency()].Value().(currency.BalanceStateValue)
		if !ok {
			return nil, mitumbase.NewBaseOperationProcessReasonError("invalid BalanceState value found, %T", gas[c.Currency()].Value()), nil
		}

		gst := state.NewStateMergeValue(gas[c.Currency()].Key(), currency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(c.Amount().Big()))))
		dst := state.NewStateMergeValue(sts[c.Currency()].Key(), currency.NewCurrencyDesignStateValue(c))
		smvs = append(smvs, gst, dst)

		sts, err := createZeroAccount(c.Currency(), getStateFunc)
		if err != nil {
			return nil, nil, err
		}

		smvs = append(smvs, sts...)
	}

	return smvs, nil, nil
}
