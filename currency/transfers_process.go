package currency

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
)

var transfersItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransfersItemProcessor)
	},
}

var transfersProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransfersProcessor)
	},
}

func (Transfers) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type TransfersItemProcessor struct {
	h    util.Hash
	item TransfersItem
	rb   map[CurrencyID]base.StateMergeValue
}

func (opp *TransfersItemProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) error {
	e := util.StringErrorFunc("failed to preprocess for TransfersItemProcessor")

	if _, err := existsState(StateKeyAccount(opp.item.Receiver()), "receiver", getStateFunc); err != nil {
		return e(err, "")
	}

	rb := map[CurrencyID]base.StateMergeValue{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		_, err := existsCurrencyPolicy(am.cid, getStateFunc)
		if err != nil {
			return err
		}

		st, _, err := getStateFunc(StateKeyBalance(opp.item.Receiver(), am.Currency()))
		if err != nil {
			return err
		}
		rb[am.Currency()] = NewBalanceStateMergeValue(st.Key(), NewBalanceStateValue(am))
	}

	opp.rb = rb

	return nil
}

func (opp *TransfersItemProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	e := util.StringErrorFunc("failed to preprocess for TransfersItemProcessor")

	sts := make([]base.StateMergeValue, len(opp.item.Amounts()))
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		v, ok := opp.rb[am.Currency()].Value().(BalanceStateValue)
		if !ok {
			return nil, e(errors.Errorf("not BalanceStateValue, %T", opp.rb[am.Currency()].Value()), "")
		}
		stv := NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(am.big)))
		sts[i] = NewBalanceStateMergeValue(opp.rb[am.Currency()].Key(), stv)
	}

	return sts, nil
}

func (opp *TransfersItemProcessor) Close() error {
	opp.h = nil
	opp.item = nil
	opp.rb = nil

	transfersItemProcessorPool.Put(opp)

	return nil
}

type TransfersProcessor struct {
	*base.BaseOperationProcessor
	sb       map[CurrencyID]base.StateMergeValue
	ns       []*TransfersItemProcessor
	required map[CurrencyID][2]Big
}

func NewTransfersProcessor() GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new TransfersProcessor")

		nopp := transfersProcessorPool.Get()
		opp, ok := nopp.(*TransfersProcessor)
		if !ok {
			return nil, e(errors.Errorf("expected TransfersProcessor, not %T", nopp), "")
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b
		opp.sb = nil
		opp.ns = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *TransfersProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(TransfersFact)
	if !ok {
		return ctx, nil, errors.Errorf("expected TransfersFact, not %T", op.Fact())
	}

	if err := checkExistsState(StateKeyAccount(fact.sender), getStateFunc); err != nil {
		return ctx, nil, err
	}

	if err := checkFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, nil, errors.Wrap(err, "invalid signing")
	}

	return ctx, nil, nil
}

func (opp *TransfersProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringErrorFunc("failed to preprocess for Transfers")

	fact, ok := op.Fact().(TransfersFact)
	if !ok {
		return nil, nil, e(errors.Errorf("expected TransfersFact, not %T", op.Fact()), "")
	}

	if required, err := opp.calculateItemsFee(op, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
	} else if sb, err := CheckEnoughBalance(fact.sender, required, getStateFunc); err != nil {
		return nil, nil, err
	} else {
		opp.required = required
		opp.sb = sb
	}

	ns := make([]*TransfersItemProcessor, len(fact.items))
	for i := range fact.items {
		cip := transfersItemProcessorPool.Get()
		c, ok := cip.(*TransfersItemProcessor)
		if !ok {
			return nil, nil, e(errors.Errorf("expected TransfersItemProcessor, not %T", cip), "")
		}

		c.h = op.Hash()
		c.item = fact.items[i]

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, nil, e(err, "")
		}

		ns[i] = c
	}
	opp.ns = ns

	var sts []base.StateMergeValue // nolint:prealloc
	for i := range opp.ns {
		s, err := opp.ns[i].Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process transfer item: %w", err), nil
		}
		sts = append(sts, s...)
	}

	for k := range opp.required {
		rq := opp.required[k]
		v, ok := opp.sb[k].Value().(BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("failed to process transfer"), nil
		}
		stv := NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(rq[0]).Sub(rq[1])))
		sts = append(sts, NewBalanceStateMergeValue(opp.sb[k].Key(), stv))
	}

	return sts, nil, nil
}

func (opp *TransfersProcessor) Close() error {
	for i := range opp.ns {
		_ = opp.ns[i].Close()
	}

	opp.sb = nil
	opp.ns = nil
	opp.required = nil

	transfersProcessorPool.Put(opp)

	return nil
}

func (opp *TransfersProcessor) calculateItemsFee(op base.Operation, getStateFunc base.GetStateFunc) (map[CurrencyID][2]Big, error) {
	fact, ok := op.Fact().(TransfersFact)
	if !ok {
		return nil, errors.Errorf("expected TransfersFact, not %T", op.Fact())
	}
	items := make([]AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return CalculateItemsFee(getStateFunc, items)
}
