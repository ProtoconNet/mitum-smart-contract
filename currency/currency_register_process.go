package currency

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
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
}

func NewCurrencyRegisterProcessor(threshold base.Threshold) GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new CurrencyRegisterProcessor")

		nopp := currencyRegisterProcessorPool.Get()
		opp, ok := nopp.(*CurrencyRegisterProcessor)
		if !ok {
			return nil, e(errors.Errorf("expected CurrencyRegisterProcessor, not %T", nopp), "")
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b
		opp.threshold = threshold

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

	nop, ok := op.(CurrencyRegister)
	if !ok {
		return ctx, nil, e(nil, "not CurrencyRegister, %T", op)
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs"), nil
	}

	fact, ok := op.Fact().(CurrencyRegisterFact)
	if !ok {
		return ctx, nil, e(nil, "not CurrencyRegisterFact, %T", op.Fact())
	}

	item := fact.currency

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

	switch _, found, err := getStateFunc(StateKeyCurrencyDesign(item.amount.Currency())); {
	case err != nil:
		return ctx, nil, err
	case found:
		return ctx, nil, base.NewBaseOperationProcessReasonError("currency already registered, %q", item.amount.Currency())
	default:
	}

	switch _, found, err := getStateFunc(StateKeyBalance(item.GenesisAccount(), item.amount.Currency())); {
	case err != nil:
		return ctx, nil, err
	case found:
		return ctx, nil, base.NewBaseOperationProcessReasonError("genesis account has already the currency, %q", item.amount.Currency())
	default:
	}

	return ctx, nil, nil
}

func (opp *CurrencyRegisterProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(CurrencyRegisterFact)
	if !ok {
		return nil, nil, errors.Errorf("not CurrencyRegisterFact, %T", op.Fact())
	}

	sts := make([]base.StateMergeValue, 4)

	item := fact.currency

	ba := NewBalanceStateValue(item.amount)
	sts[0] = NewBalanceStateMergeValue(
		StateKeyBalance(item.genesisAccount, item.amount.cid),
		ba,
	)

	de := NewCurrencyDesignStateValue(item)
	sts[1] = NewCurrencyDesignStateMergeValue(StateKeyCurrencyDesign(item.amount.cid), de)

	{
		l, err := createZeroAccount(item.amount.Currency(), getStateFunc)
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

	currencyRegisterProcessorPool.Put(opp)

	return nil
}
