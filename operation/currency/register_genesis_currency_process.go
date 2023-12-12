package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

func (op RegisterGenesisCurrency) PreProcess(
	ctx context.Context, _ base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	return ctx, nil, nil
}

func (op RegisterGenesisCurrency) Process(
	_ context.Context, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(RegisterGenesisCurrencyFact)
	if !ok {
		return nil, nil, errors.Errorf("expected %T, not %T", RegisterGenesisCurrencyFact{}, op.Fact())
	}

	newAddress, err := fact.Address()
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to get genesis account address, %w", err), nil
	}

	ns, err := state.NotExistsState(currency.StateKeyAccount(newAddress), "key of genesis", getStateFunc)
	if err != nil {
		return nil, nil, err
	}

	cs := make([]types.CurrencyDesign, len(fact.cs))
	gas := map[types.CurrencyID]base.StateMergeValue{}
	sts := map[types.CurrencyID]base.StateMergeValue{}
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
		//gas[c.Currency()] = state.NewStateMergeValue(st.Key(), currency.NewBalanceStateValue(types.NewZeroAmount(c.Currency())))

		gas[c.Currency()] = common.NewBaseStateMergeValue(
			st.Key(),
			currency.NewAddBalanceStateValue(types.NewZeroAmount(c.Currency())),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currency.NewBalanceStateValueMerger(
					height,
					st.Key(),
					c.Currency(),
					st,
				)
			},
		)
	}

	var smvs []base.StateMergeValue
	if ac, err := types.NewAccount(newAddress, fact.keys); err != nil {
		return nil, nil, err
	} else {
		smvs = append(smvs, state.NewStateMergeValue(ns.Key(), currency.NewAccountStateValue(ac)))
	}

	for i := range cs {
		c := cs[i]
		v, ok := gas[c.Currency()].Value().(currency.AddBalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("invalid State value found, %T", gas[c.Currency()].Value()), nil
		}

		gst := common.NewBaseStateMergeValue(
			gas[c.Currency()].Key(),
			currency.NewAddBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(c.Amount().Big()))),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currency.NewBalanceStateValueMerger(
					height,
					gas[c.Currency()].Key(),
					c.Currency(),
					st,
				)
			},
		)

		//gst := state.NewStateMergeValue(gas[c.Currency()].Key(), currency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(c.Amount().Big()))))
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
