package currency

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var currencyPolicyUpdaterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CurrencyPolicyUpdaterProcessor)
	},
}

func (CurrencyPolicy) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type CurrencyPolicyUpdaterProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewCurrencyPolicyUpdaterProcessor(threshold base.Threshold) GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new CurrencyPolicyUpdaterProcessor")

		nopp := currencyPolicyUpdaterProcessorPool.Get()
		opp, ok := nopp.(*CurrencyPolicyUpdaterProcessor)
		if !ok {
			return nil, e(errors.Errorf("expected CurrencyPolicyUpdaterProcessor, not %T", nopp), "")
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

func (opp *CurrencyPolicyUpdaterProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess for CurrencyPolicyUpdater")

	nop, ok := op.(CurrencyPolicyUpdater)
	if !ok {
		return ctx, nil, e(nil, "not CurrencyPolicyUpdater, %T", op)
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs"), nil
	}

	fact, ok := op.Fact().(CurrencyPolicyUpdaterFact)
	if !ok {
		return ctx, nil, e(nil, "not CurrencyPolicyUpdaterFact, %T", op.Fact())
	}

	err := checkExistsState(StateKeyCurrencyDesign(fact.currency), getStateFunc)
	if err != nil {
		return ctx, nil, err
	}

	if receiver := fact.policy.Feeer().Receiver(); receiver != nil {
		if err := checkExistsState(StateKeyAccount(receiver), getStateFunc); err != nil {
			return ctx, nil, e(err, "feeer receiver account not found")
		}
	}

	if err := checkExistsState(StateKeyCurrencyDesign(fact.currency), getStateFunc); err != nil {
		return ctx, nil, base.NewBaseOperationProcessReasonError("currency not found, %q", fact.currency)
	}

	return ctx, nil, nil
}

func (opp *CurrencyPolicyUpdaterProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(CurrencyPolicyUpdaterFact)
	if !ok {
		return nil, nil, errors.Errorf("not CurrencyPolicyUpdaterFact, %T", op.Fact())
	}

	sts := make([]base.StateMergeValue, 1)

	st, err := existsState(StateKeyCurrencyDesign(fact.currency), "currency design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of currency %w : %v", fact.currency, err), nil
	}

	de, err := StateCurrencyDesignValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to get currency design of %w : %v", fact.currency, err), nil
	}

	de.policy = fact.policy

	c := NewCurrencyDesignStateMergeValue(
		st.Key(),
		NewCurrencyDesignStateValue(de),
	)
	sts[0] = c

	return sts, nil, nil
}

func (opp *CurrencyPolicyUpdaterProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0

	currencyPolicyUpdaterProcessorPool.Put(opp)

	return nil
}
