package currency

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"

	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
)

var mintProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(MintProcessor)
	},
}

type MintProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewMintProcessor(threshold base.Threshold) types.GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("create new MintProcessor")

		nopp := mintProcessorPool.Get()
		opp, ok := nopp.(*MintProcessor)
		if !ok {
			return nil, e.Errorf("expected MintProcessor, not %T", nopp)
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
				return nil, e.Wrap(isaac.ErrStopProcessingRetry.Errorf("get suffrage from state"))
			}

			opp.suffrage = suf
		}

		return opp, nil
	}
}

func (opp *MintProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("preprocess Mint")

	nop, ok := op.(Mint)
	if !ok {
		return ctx, nil, e.Errorf("expected Mint, not %T", op)
	}

	fact, ok := op.Fact().(MintFact)
	if !ok {
		return ctx, nil, e.Errorf("expected MintFact, not %T", op.Fact())
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs; %w", err), nil
	}

	for i := range fact.Items() {
		item := fact.Items()[i]

		err := state.CheckExistsState(currency.StateKeyCurrencyDesign(item.Amount().Currency()), getStateFunc)
		if err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("currency not found, %v; %v", item.Amount().Currency(), err.Error()), nil
		}

		err = state.CheckExistsState(currency.StateKeyAccount(item.Receiver()), getStateFunc)
		if err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("receiver not found, %v; %v", item.Receiver(), err.Error()), nil
		}

		err = state.CheckNotExistsState(extension.StateKeyContractAccount(item.Receiver()), getStateFunc)
		if err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("contract account cannot be mint receiver, %v; %v", item.Receiver(), err.Error()), nil
		}
	}

	return ctx, nil, nil
}

func (opp *MintProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("process Mint")

	fact, ok := op.Fact().(MintFact)
	if !ok {
		return nil, nil, e.Errorf("expected MintFact, not %T", op.Fact())
	}

	var sts []base.StateMergeValue

	aggs := map[types.CurrencyID]common.Big{}

	for i := range fact.Items() {
		item := fact.Items()[i]

		var ab types.Amount

		k := currency.StateKeyBalance(item.Receiver(), item.Amount().Currency())
		switch st, found, err := getStateFunc(k); {
		case err != nil:
			return nil, base.NewBaseOperationProcessReasonError("find receiver balance state, %v; %w", k, err), nil
		case !found:
			ab = types.NewZeroAmount(item.Amount().Currency())
		default:
			b, err := currency.StateBalanceValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("get balance value, %v; %w", k, err), nil
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
			return nil, base.NewBaseOperationProcessReasonError("find currency design state, %v %w", cid, err), nil
		case !found:
			return nil, base.NewBaseOperationProcessReasonError("currency not found, %v; %w", cid, err), nil
		default:
			d, err := currency.StateCurrencyDesignValue(st)
			if err != nil {
				return nil, base.NewBaseOperationProcessReasonError("get currency design value, %v; %w", cid, err), nil
			}
			de = d
		}

		ade, err := de.AddAggregate(big)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("add aggregate, %v; %w", cid, err), nil
		}

		sts = append(sts, state.NewStateMergeValue(k, currency.NewCurrencyDesignStateValue(ade)))
	}

	return sts, nil, nil
}

func (opp *MintProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0

	mintProcessorPool.Put(opp)

	return nil
}
