package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v2/base"
	types "github.com/ProtoconNet/mitum-currency/v2/operation/type"
	"github.com/ProtoconNet/mitum-currency/v2/state"
	"github.com/ProtoconNet/mitum-currency/v2/state/currency"
	"github.com/ProtoconNet/mitum-currency/v2/state/extension"
	"sync"

	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
)

var suffrageInflationProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(SuffrageInflationProcessor)
	},
}

type SuffrageInflationProcessor struct {
	*mitumbase.BaseOperationProcessor
	suffrage  mitumbase.Suffrage
	threshold mitumbase.Threshold
}

func NewSuffrageInflationProcessor(threshold mitumbase.Threshold) types.GetNewProcessor {
	return func(height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new SuffrageInflationProcessor")

		nopp := suffrageInflationProcessorPool.Get()
		opp, ok := nopp.(*SuffrageInflationProcessor)
		if !ok {
			return nil, e(nil, "expected SuffrageInflationProcessor, not %T", nopp)
		}

		b, err := mitumbase.NewBaseOperationProcessor(
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
			sufstv := i.Value().(mitumbase.SuffrageNodesStateValue) //nolint:forcetypeassert //...

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
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess SuffrageInflation")

	nop, ok := op.(SuffrageInflation)
	if !ok {
		return ctx, nil, e(nil, "expected SuffrageInflation, not %T", op)
	}

	fact, ok := op.Fact().(SuffrageInflationFact)
	if !ok {
		return ctx, nil, e(nil, "expected SuffrageInflationFact, not %T", op.Fact())
	}

	if err := mitumbase.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("not enough signs: %w", err), nil
	}

	for i := range fact.Items() {
		item := fact.Items()[i]

		err := state.CheckExistsState(currency.StateKeyCurrencyDesign(item.Amount().Currency()), getStateFunc)
		if err != nil {
			return ctx, mitumbase.NewBaseOperationProcessReasonError("currency not found, %q: %w", item.Amount().Currency(), err.Error()), nil
		}

		err = state.CheckExistsState(currency.StateKeyAccount(item.Receiver()), getStateFunc)
		if err != nil {
			return ctx, mitumbase.NewBaseOperationProcessReasonError("receiver not found, %q: %w", item.Receiver(), err.Error()), nil
		}

		err = state.CheckNotExistsState(extension.StateKeyContractAccount(item.Receiver()), getStateFunc)
		if err != nil {
			return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account cannot be suffrage-inflation receiver, %q: %w", item.Receiver(), err.Error()), nil
		}
	}

	return ctx, nil, nil
}

func (opp *SuffrageInflationProcessor) Process(
	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringErrorFunc("failed to process SuffrageInflation")

	fact, ok := op.Fact().(SuffrageInflationFact)
	if !ok {
		return nil, nil, e(nil, "expected SuffrageInflationFact, not %T", op.Fact())
	}

	var sts []mitumbase.StateMergeValue

	aggs := map[base.CurrencyID]base.Big{}

	for i := range fact.Items() {
		item := fact.Items()[i]

		var ab base.Amount

		k := currency.StateKeyBalance(item.Receiver(), item.Amount().Currency())
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to find receiver balance state, %q: %w", k, err), nil
		case !found:
			ab = base.NewZeroAmount(item.Amount().Currency())
		default:
			b, err := currency.StateBalanceValue(st)
			if err != nil {
				return nil, mitumbase.NewBaseOperationProcessReasonError("failed to get balance value, %q: %w", k, err), nil
			}
			ab = b
		}

		sts = append(sts, state.NewStateMergeValue(k, currency.NewBalanceStateValue(base.NewAmount(ab.Big().Add(item.Amount().Big()), item.Amount().Currency()))))

		if _, found := aggs[item.Amount().Currency()]; found {
			aggs[item.Amount().Currency()] = aggs[item.Amount().Currency()].Add(item.Amount().Big())
		} else {
			aggs[item.Amount().Currency()] = item.Amount().Big()
		}
	}

	for cid, big := range aggs {
		var de base.CurrencyDesign

		k := currency.StateKeyCurrencyDesign(cid)
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to find currency design state, %q: %w", cid, err), nil
		case !found:
			return nil, mitumbase.NewBaseOperationProcessReasonError("currency not found, %q: %w", cid, err), nil
		default:
			d, err := currency.StateCurrencyDesignValue(st)
			if err != nil {
				return nil, mitumbase.NewBaseOperationProcessReasonError("failed to get currency design value, %q: %w", cid, err), nil
			}
			de = d
		}

		ade, err := de.AddAggregate(big)
		if err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to add aggregate, %q: %w", cid, err), nil
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
