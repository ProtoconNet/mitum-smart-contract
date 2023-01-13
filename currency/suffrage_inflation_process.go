package currency

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/util"
)

var suffrageInflationProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(SuffrageInflationProcessor)
	},
}

func (SuffrageInflation) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type SuffrageInflationProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewSuffrageInflationProcessor(threshold base.Threshold) GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new SuffrageInflationProcessor")

		nopp := suffrageInflationProcessorPool.Get()
		opp, ok := nopp.(*SuffrageInflationProcessor)
		if !ok {
			return nil, e(errors.Errorf("expected SuffrageInflationProcessor, not %T", nopp), "")
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

func (opp *SuffrageInflationProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess for SuffrageInflation")

	nop, ok := op.(SuffrageInflation)
	if !ok {
		return ctx, nil, e(nil, "not SuffrageInflation, %T", op)
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs"), nil
	}

	fact, ok := op.Fact().(SuffrageInflationFact)
	if !ok {
		return ctx, nil, e(nil, "not SuffrageInflationFact, %T", op.Fact())
	}

	for i := range fact.items {
		item := fact.items[i]

		err := checkExistsState(StateKeyCurrencyDesign(item.amount.cid), getStateFunc)
		if err != nil {
			return ctx, nil, err
		}

		err = checkExistsState(StateKeyAccount(item.receiver), getStateFunc)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, nil, nil
}

func (opp *SuffrageInflationProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(SuffrageInflationFact)
	if !ok {
		return nil, nil, errors.Errorf("not SuffrageInflationFact, %T", op.Fact())
	}

	sts := []base.StateMergeValue{}

	aggs := map[CurrencyID]Big{}

	for i := range fact.items {
		item := fact.items[i]

		var ab Amount

		k := StateKeyBalance(item.receiver, item.amount.cid)
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError("failed to find balance state %v: %w", k, err), nil
		case !found:
			ab = NewZeroAmount(item.amount.cid)
		default:
			b, err := StateBalanceValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get balance %v: %w", k, err), nil
			}
			ab = b
		}

		sts = append(sts, NewBalanceStateMergeValue(k, NewBalanceStateValue(NewAmount(ab.big.Add(item.amount.big), item.amount.cid))))

		if _, found := aggs[item.amount.cid]; found {
			aggs[item.amount.cid] = aggs[item.amount.cid].Add(item.amount.big)
		} else {
			aggs[item.amount.cid] = item.amount.big
		}
	}

	for cid, big := range aggs {
		var de CurrencyDesign

		k := StateKeyCurrencyDesign(cid)
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError("failed to find currency design state %v: %w", cid, err), nil
		case !found:
			return nil, base.NewBaseOperationProcessReasonError("currency design not found %v: %w", cid, err), nil
		default:
			d, err := StateCurrencyDesignValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get currency design %v: %w", cid, err), nil
			}
			de = d
		}

		ade, err := de.AddAggregate(big)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to add aggregate, %v: %w", cid, err), nil
		}

		sts = append(sts, NewCurrencyDesignStateMergeValue(k, NewCurrencyDesignStateValue(ade)))
	}

	return sts, nil, nil
}

func (opp *SuffrageInflationProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0

	suffrageInflationProcessorPool.Put(opp)

	return nil
}
