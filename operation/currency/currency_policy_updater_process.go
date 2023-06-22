package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
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

func (CurrencyPolicyUpdater) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type CurrencyPolicyUpdaterProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewCurrencyPolicyUpdaterProcessor(threshold base.Threshold) types.GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new CurrencyPolicyUpdaterProcessor")

		nopp := currencyPolicyUpdaterProcessorPool.Get()
		opp, ok := nopp.(*CurrencyPolicyUpdaterProcessor)
		if !ok {
			return nil, e.Wrap(errors.Errorf("expected CurrencyPolicyUpdaterProcessor, not %T", nopp))
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b
		opp.threshold = threshold

		switch i, found, err := getStateFunc(isaac.SuffrageStateKey); {
		case err != nil:
			return nil, e.Wrap(err)
		case !found, i == nil:
			return nil, e.Wrap(isaac.ErrStopProcessingRetry.Errorf("empty state"))
		default:
			sufstv := i.Value().(base.SuffrageNodesStateValue) //nolint:forcetypeassert //...

			suf, err := sufstv.Suffrage()
			if err != nil {
				return nil, e.Wrap(isaac.ErrStopProcessingRetry.Errorf("failed to get suffrage from state"))
			}

			opp.suffrage = suf
		}

		return opp, nil
	}
}

func (opp *CurrencyPolicyUpdaterProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess for CurrencyPolicyUpdater")

	nop, ok := op.(CurrencyPolicyUpdater)
	if !ok {
		return ctx, nil, e.Errorf("not CurrencyPolicyUpdater, %T", op)
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs"), nil
	}

	fact, ok := op.Fact().(CurrencyPolicyUpdaterFact)
	if !ok {
		return ctx, nil, e.Errorf("not CurrencyPolicyUpdaterFact, %T", op.Fact())
	}

	err := state.CheckExistsState(statecurrency.StateKeyCurrencyDesign(fact.currency), getStateFunc)
	if err != nil {
		return ctx, nil, err
	}

	if receiver := fact.policy.Feeer().Receiver(); receiver != nil {
		if err := state.CheckExistsState(statecurrency.StateKeyAccount(receiver), getStateFunc); err != nil {
			return ctx, nil, e.WithMessage(err, "feeer receiver account not found")
		}
	}

	if err := state.CheckExistsState(statecurrency.StateKeyCurrencyDesign(fact.currency), getStateFunc); err != nil {
		return ctx, nil, base.NewBaseOperationProcessReasonError("currency not found, %q", fact.currency)
	}

	return ctx, nil, nil
}

func (opp *CurrencyPolicyUpdaterProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(CurrencyPolicyUpdaterFact)
	if !ok {
		return nil, nil, errors.Errorf("not CurrencyPolicyUpdaterFact, %T", op.Fact())
	}

	sts := make([]base.StateMergeValue, 1)

	st, err := state.ExistsState(statecurrency.StateKeyCurrencyDesign(fact.currency), "currency design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of currency %v : %v", fact.currency, err), nil
	}

	de, err := statecurrency.StateCurrencyDesignValue(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to get currency design of %v : %v", fact.currency, err), nil
	}

	de.SetPolicy(fact.policy)

	c := state.NewStateMergeValue(
		st.Key(),
		statecurrency.NewCurrencyDesignStateValue(de),
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
