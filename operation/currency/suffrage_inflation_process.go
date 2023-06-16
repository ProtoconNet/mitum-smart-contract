package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"sync"

	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
)

var suffrageInflationProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(SuffrageInflationProcessor)
	},
}

type SuffrageInflationProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewSuffrageInflationProcessor(threshold base.Threshold) types.GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new SuffrageInflationProcessor")

		nopp := suffrageInflationProcessorPool.Get()
		opp, ok := nopp.(*SuffrageInflationProcessor)
		if !ok {
			return nil, e.Errorf("expected SuffrageInflationProcessor, not %T", nopp)
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

func (opp *SuffrageInflationProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess SuffrageInflation")

	nop, ok := op.(SuffrageInflation)
	if !ok {
		return ctx, nil, e.Errorf("expected SuffrageInflation, not %T", op)
	}

	fact, ok := op.Fact().(SuffrageInflationFact)
	if !ok {
		return ctx, nil, e.Errorf("expected SuffrageInflationFact, not %T", op.Fact())
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs: %w", err), nil
	}

	for i := range fact.Items() {
		item := fact.Items()[i]

		err := state.CheckExistsState(currency.StateKeyCurrencyDesign(item.Amount().Currency()), getStateFunc)
		if err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("currency not found, %q: %w", item.Amount().Currency(), err.Error()), nil
		}

		err = state.CheckExistsState(currency.StateKeyAccount(item.Receiver()), getStateFunc)
		if err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("receiver not found, %q: %w", item.Receiver(), err.Error()), nil
		}

		err = state.CheckNotExistsState(extension.StateKeyContractAccount(item.Receiver()), getStateFunc)
		if err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("contract account cannot be suffrage-inflation receiver, %q: %w", item.Receiver(), err.Error()), nil
		}
	}

	return ctx, nil, nil
}

func (opp *SuffrageInflationProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process SuffrageInflation")

	fact, ok := op.Fact().(SuffrageInflationFact)
	if !ok {
		return nil, nil, e.Errorf("expected SuffrageInflationFact, not %T", op.Fact())
	}

	var sts []base.StateMergeValue

	aggs := map[types.CurrencyID]common.Big{}

	for i := range fact.Items() {
		item := fact.Items()[i]

		var ab types.Amount

		k := currency.StateKeyBalance(item.Receiver(), item.Amount().Currency())
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError("failed to find receiver balance state, %q: %w", k, err), nil
		case !found:
			ab = types.NewZeroAmount(item.Amount().Currency())
		default:
			b, err := currency.StateBalanceValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get balance value, %q: %w", k, err), nil
			}
			ab = b
		}

		sts = append(sts, state.NewStateMergeValue(k, currency.NewBalanceStateValue(types.NewAmount(ab.Big().Add(item.Amount().Big()), item.Amount().Currency()))))

		if _, found := aggs[item.Amount().Currency()]; found {
			aggs[item.Amount().Currency()] = aggs[item.Amount().Currency()].Add(item.Amount().Big())
		} else {
			aggs[item.Amount().Currency()] = item.Amount().Big()
		}
	}

	for cid, big := range aggs {
		var de types.CurrencyDesign

		k := currency.StateKeyCurrencyDesign(cid)
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError("failed to find currency design state, %q: %w", cid, err), nil
		case !found:
			return nil, base.NewBaseOperationProcessReasonError("currency not found, %q: %w", cid, err), nil
		default:
			d, err := currency.StateCurrencyDesignValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("failed to get currency design value, %q: %w", cid, err), nil
			}
			de = d
		}

		ade, err := de.AddAggregate(big)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to add aggregate, %q: %w", cid, err), nil
		}

		sts = append(sts, state.NewStateMergeValue(k, currency.NewCurrencyDesignStateValue(ade)))
	}

	return sts, nil, nil
}

func (opp *SuffrageInflationProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0

	suffrageInflationProcessorPool.Put(opp)

	return nil
}
