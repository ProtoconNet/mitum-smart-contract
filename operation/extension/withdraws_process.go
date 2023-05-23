package extension

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum-currency/v2/operation/currency"
	types "github.com/ProtoconNet/mitum-currency/v2/operation/type"
	"github.com/ProtoconNet/mitum-currency/v2/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v2/state/currency"
	"github.com/ProtoconNet/mitum-currency/v2/state/extension"
	"sync"

	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var withdrawsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawsItemProcessor)
	},
}

var withdrawsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawsProcessor)
	},
}

func (Withdraws) Process(
	_ context.Context, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type WithdrawsItemProcessor struct {
	h      util.Hash
	sender mitumbase.Address
	item   WithdrawsItem
	tb     map[base.CurrencyID]mitumbase.StateMergeValue
}

func (opp *WithdrawsItemProcessor) PreProcess(
	_ context.Context, _ mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) error {
	if err := state.CheckExistsState(statecurrency.StateKeyAccount(opp.item.Target()), getStateFunc); err != nil {
		return err
	}

	st, err := state.ExistsState(extension.StateKeyContractAccount(opp.item.Target()), "key of target contract account", getStateFunc)
	if err != nil {
		return err
	}
	v, err := extension.StateContractAccountValue(st)
	if err != nil {
		return err
	}
	if !v.Owner().Equal(opp.sender) {
		return errors.Errorf("contract account owner is not matched with %q", opp.sender)
	}

	tb := map[base.CurrencyID]mitumbase.StateMergeValue{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		_, err := state.ExistsCurrencyPolicy(am.Currency(), getStateFunc)
		if err != nil {
			return err
		}

		st, _, err := getStateFunc(statecurrency.StateKeyBalance(opp.item.Target(), am.Currency()))
		if err != nil {
			return err
		}

		balance, err := statecurrency.StateBalanceValue(st)
		if err != nil {
			return err
		}

		tb[am.Currency()] = state.NewStateMergeValue(st.Key(), statecurrency.NewBalanceStateValue(balance))
	}

	opp.tb = tb

	return nil
}

func (opp *WithdrawsItemProcessor) Process(
	_ context.Context, _ mitumbase.Operation, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, error) {
	sts := make([]mitumbase.StateMergeValue, len(opp.item.Amounts()))
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		v, ok := opp.tb[am.Currency()].Value().(statecurrency.BalanceStateValue)
		if !ok {
			return nil, errors.Errorf("expect BalanceStateValue, not %T", opp.tb[am.Currency()].Value())
		}
		stv := statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(am.Big())))
		sts[i] = state.NewStateMergeValue(opp.tb[am.Currency()].Key(), stv)
	}

	return sts, nil
}

func (opp *WithdrawsItemProcessor) Close() {
	opp.h = nil
	opp.sender = nil
	opp.item = nil
	opp.tb = nil

	withdrawsItemProcessorPool.Put(opp)
}

type WithdrawsProcessor struct {
	*mitumbase.BaseOperationProcessor
	ns       []*WithdrawsItemProcessor
	required map[base.CurrencyID][2]base.Big // required[0] : amount + fee, required[1] : fee
}

func NewWithdrawsProcessor() types.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new WithdrawsProcessor")

		nopp := withdrawsProcessorPool.Get()
		opp, ok := nopp.(*WithdrawsProcessor)
		if !ok {
			return nil, e(nil, "expected WithdrawsProcessor, not %T", nopp)
		}

		b, err := mitumbase.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *WithdrawsProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess Withdraws")

	fact, ok := op.Fact().(WithdrawsFact)
	if !ok {
		return ctx, nil, e(nil, "expected WithdrawsFact, not %T", op.Fact())
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.sender), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.sender, err), nil
	}

	if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.sender), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account cannot be ca withdraw sender, %q: %w", fact.sender, err), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	return ctx, nil, nil
}

func (opp *WithdrawsProcessor) Process( // nolint:dupl
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(WithdrawsFact)
	if !ok {
		return nil, mitumbase.NewBaseOperationProcessReasonError("expected WithdrawsFact, not %T", op.Fact()), nil
	}

	feeReceiveBalSts, required, err := opp.calculateItemsFee(op, getStateFunc)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
	}
	senderBalSts, err := currency.CheckEnoughBalance(fact.sender, required, getStateFunc)
	if err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check enough balance: %w", err), nil
	} else {
		opp.required = required
	}

	ns := make([]*WithdrawsItemProcessor, len(fact.items))
	for i := range fact.items {
		cip := withdrawsItemProcessorPool.Get()
		c, ok := cip.(*WithdrawsItemProcessor)
		if !ok {
			return nil, mitumbase.NewBaseOperationProcessReasonError("expected WithdrawsItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.sender = fact.sender
		c.item = fact.items[i]

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("fail to preprocess WithdrawsItem: %w", err), nil
		}

		ns[i] = c
	}

	var stateMergeValues []mitumbase.StateMergeValue // nolint:prealloc
	for i := range ns {
		s, err := ns[i].Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to process WithdrawsItem: %w", err), nil
		}
		stateMergeValues = append(stateMergeValues, s...)

		ns[i].Close()
	}

	for cid := range senderBalSts {
		v, ok := senderBalSts[cid].Value().(statecurrency.BalanceStateValue)
		if !ok {
			return nil, mitumbase.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", senderBalSts[cid].Value()), nil
		}

		var stateMergeValue mitumbase.StateMergeValue
		if senderBalSts[cid].Key() == feeReceiveBalSts[cid].Key() {
			stateMergeValue = state.NewStateMergeValue(
				senderBalSts[cid].Key(),
				statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[cid][0]).Add(opp.required[cid][1]))),
			)
		} else {
			stateMergeValue = state.NewStateMergeValue(
				senderBalSts[cid].Key(),
				statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[cid][0]))),
			)
			r, ok := feeReceiveBalSts[cid].Value().(statecurrency.BalanceStateValue)
			if !ok {
				return nil, mitumbase.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", feeReceiveBalSts[cid].Value()), nil
			}
			stateMergeValues = append(
				stateMergeValues,
				state.NewStateMergeValue(
					feeReceiveBalSts[cid].Key(),
					statecurrency.NewBalanceStateValue(r.Amount.WithBig(r.Amount.Big().Add(opp.required[cid][1]))),
				),
			)
		}
		stateMergeValues = append(stateMergeValues, stateMergeValue)
	}

	return stateMergeValues, nil, nil
}

func (opp *WithdrawsProcessor) Close() error {
	for i := range opp.ns {
		opp.ns[i].Close()
	}

	opp.required = nil
	withdrawsProcessorPool.Put(opp)

	return nil
}

func (opp *WithdrawsProcessor) calculateItemsFee(op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (map[base.CurrencyID]mitumbase.State, map[base.CurrencyID][2]base.Big, error) {
	fact, ok := op.Fact().(WithdrawsFact)
	if !ok {
		return nil, nil, errors.Errorf("expected WithdrawsFact, not %T", op.Fact())
	}
	items := make([]currency.AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return currency.CalculateItemsFee(getStateFunc, items)
}
