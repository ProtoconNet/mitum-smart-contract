package currency

import (
	"context"
	"sync"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/util"
)

var currencyRegisterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CurrencyRegisterProcessor)
	},
}

func (CurrencyRegister) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type CurrencyRegisterProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
	ga        base.StateMergeValue
	de        base.StateMergeValue
}

func NewCurrencyRegisterProcessor(threshold base.Threshold) GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new CurrencyRegisterProcessor")

		opp := currencyRegisterProcessorPool.Get().(*CurrencyRegisterProcessor)
		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b
		opp.threshold = threshold
		opp.ga = nil
		opp.de = nil

		switch i, found, err := getStateFunc(isaac.SuffrageStateKey); {
		case err != nil:
			return nil, e(err, "")
		case !found, i == nil:
			return nil, e(isaac.ErrStopProcessingRetry.Errorf("empty state"), "")
		default:
			sufstv := i.Value().(base.SuffrageNodesStateValue) //nolint:forcetypeassert //...

			suf, err := sufstv.Suffrage()
			if err != nil {
				return nil, e(isaac.ErrStopProcessingRetry.Errorf("failed to get suffrage from state"), "")
			}

			opp.suffrage = suf
		}

		return opp, nil
	}
}

func (opp *CurrencyRegisterProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess for CurrencyRegister")

	noop, ok := op.(CurrencyRegister)
	if !ok {
		return ctx, nil, e(nil, "not CurrencyRegister, %T", op)
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, noop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs"), nil
	}

	item := op.Fact().(CurrencyRegisterFact).currency

	_, err := notExistsState(StateKeyCurrencyDesign(item.amount.Currency()), item.amount.Currency().String(), getStateFunc)
	if err != nil {
		return ctx, nil, err
	}

	if err := checkExistsState(StateKeyAccount(item.GenesisAccount()), getStateFunc); err != nil {
		return ctx, nil, e(err, "genesis account not found")
	}

	if receiver := item.Policy().Feeer().Receiver(); receiver != nil {
		if err := checkExistsState(StateKeyAccount(receiver), getStateFunc); err != nil {
			return ctx, nil, e(err, "feeer receiver account not found")
		}
	}

	switch st, found, err := getStateFunc(StateKeyCurrencyDesign(item.amount.Currency())); {
	case err != nil:
		return ctx, nil, err
	case found:
		return ctx, nil, base.NewBaseOperationProcessReasonError("currency already registered, %q", item.amount.Currency())
	default:
		opp.de = NewCurrencyDesignStateMergeValue(st.Key(), st.Value())
	}

	switch st, found, err := getStateFunc(StateKeyBalance(item.GenesisAccount(), item.amount.Currency())); {
	case err != nil:
		return ctx, nil, err
	case found:
		return ctx, nil, base.NewBaseOperationProcessReasonError("genesis account has already the currency, %q", item.amount.Currency())
	default:
		opp.ga = NewBalanceStateMergeValue(st.Key(), NewBalanceStateValue(NewZeroAmount(item.amount.Currency())))
	}

	return ctx, nil, nil
}

func (opp *CurrencyRegisterProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact := op.Fact().(CurrencyRegisterFact)

	sts := make([]base.StateMergeValue, 4)
	v, ok := opp.ga.Value().(BalanceStateValue)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("invalid BalanceStateValue found, %T", opp.ga.Value()), nil
	}
	sts[0] = NewBalanceStateMergeValue(
		opp.ga.Key(),
		NewBalanceStateValue(v.Amount.WithBig(v.Amount.big.Add(fact.currency.amount.Big()))),
	)
	c := NewCurrencyDesignStateMergeValue(
		opp.de.Key(),
		NewCurrencyDesignStateValue(fact.currency),
	)
	sts[1] = c

	{
		l, err := createZeroAccount(fact.currency.amount.Currency(), getStateFunc)
		if err != nil {
			return nil, nil, err
		}
		sts[2], sts[3] = l[0], l[1]
	}

	return sts, nil, nil
}

func createZeroAccount(
	cid CurrencyID,
	getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	sts := make([]base.StateMergeValue, 2)

	ac, err := ZeroAccount(cid)
	if err != nil {
		return nil, err
	}
	ast, err := notExistsState(StateKeyAccount(ac.Address()), "keys of zero account", getStateFunc)
	if err != nil {
		return nil, err
	}

	sts[0] = NewAccountStateMergeValue(ast.Key(), NewAccountStateValue(ac))

	bst, err := notExistsState(StateKeyBalance(ac.Address(), cid), "balance of zero account", getStateFunc)
	if err != nil {
		return nil, err
	}

	sts[1] = NewBalanceStateMergeValue(bst.Key(), NewBalanceStateValue(NewZeroAmount(cid)))

	return sts, nil
}

func (opp *CurrencyRegisterProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0
	opp.ga = nil
	opp.de = nil

	currencyRegisterProcessorPool.Put(opp)

	return nil
}
