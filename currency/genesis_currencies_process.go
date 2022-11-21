package currency

import (
	"context"

	"github.com/spikeekips/mitum/base"
)

func (_ GenesisCurrencies) PreProcess(
	ctx context.Context, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	return ctx, nil, nil
}

func (op GenesisCurrencies) Process(
	ctx context.Context, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact := op.Fact().(GenesisCurrenciesFact)

	newAddress, err := fact.Address()
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(err.Error()), nil
	}

	ns, err := notExistsState(StateKeyAccount(newAddress), "key of genesis", getStateFunc)
	if err != nil {
		return nil, nil, err
	}

	cs := make([]CurrencyDesign, len(fact.cs))
	gas := map[CurrencyID]base.StateMergeValue{}
	sts := map[CurrencyID]base.StateMergeValue{}
	for i := range fact.cs {
		c := fact.cs[i]
		c.genesisAccount = newAddress
		cs[i] = c

		st, err := notExistsState(StateKeyCurrencyDesign(c.amount.Currency()), "currency", getStateFunc)
		if err != nil {
			return nil, nil, err
		}

		sts[c.amount.Currency()] = NewCurrencyDesignStateMergeValue(st.Key(), NewCurrencyDesignStateValue(c))

		st, err = notExistsState(StateKeyBalance(newAddress, c.amount.Currency()), "balance of genesis", getStateFunc)
		if err != nil {
			return nil, nil, err
		}
		gas[c.amount.Currency()] = NewBalanceStateMergeValue(st.Key(), NewBalanceStateValue(NewZeroAmount(c.amount.Currency())))
	}

	var smvs []base.StateMergeValue
	if ac, err := NewAccount(newAddress, fact.keys); err != nil {
		return nil, nil, err
	} else {
		smvs = append(smvs, NewAccountStateMergeValue(ns.Key(), NewAccountStateValue(ac)))
	}

	for i := range cs {
		c := cs[i]
		v, ok := gas[c.amount.Currency()].Value().(BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("invalid BalanceState value found, %T", gas[c.amount.Currency()].Value()), nil
		}
		gst := NewBalanceStateMergeValue(gas[c.amount.Currency()].Key(), NewBalanceStateValue(v.Amount.WithBig(v.Amount.big.Add(c.amount.Big()))))
		dst := NewCurrencyDesignStateMergeValue(sts[c.amount.Currency()].Key(), NewCurrencyDesignStateValue(c))
		smvs = append(smvs, gst, dst)

		sts, err := createZeroAccount(c.amount.Currency(), getStateFunc)
		if err != nil {
			return nil, nil, err
		}

		smvs = append(smvs, sts...)
	}

	return smvs, nil, nil
}
