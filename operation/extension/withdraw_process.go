package extension

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var withdrawItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawItemProcessor)
	},
}

var withdrawProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(WithdrawProcessor)
	},
}

func (Withdraw) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type WithdrawItemProcessor struct {
	h      util.Hash
	sender base.Address
	item   WithdrawItem
	tb     map[types.CurrencyID]base.StateMergeValue
}

func (opp *WithdrawItemProcessor) PreProcess(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
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
		return errors.Errorf("contract account owner is not matched with %v", opp.sender)
	}

	tb := map[types.CurrencyID]base.StateMergeValue{}
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

func (opp *WithdrawItemProcessor) Process(
	_ context.Context, _ base.Operation, _ base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	sts := make([]base.StateMergeValue, len(opp.item.Amounts()))
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

func (opp *WithdrawItemProcessor) Close() {
	opp.h = nil
	opp.sender = nil
	opp.item = nil
	opp.tb = nil

	withdrawItemProcessorPool.Put(opp)
}

type WithdrawProcessor struct {
	*base.BaseOperationProcessor
	ns       []*WithdrawItemProcessor
	required map[types.CurrencyID][2]common.Big // required[0] : amount + fee, required[1] : fee
}

func NewWithdrawProcessor() types.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new WithdrawProcessor")

		nopp := withdrawProcessorPool.Get()
		opp, ok := nopp.(*WithdrawProcessor)
		if !ok {
			return nil, e.WithMessage(nil, "expected WithdrawProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *WithdrawProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess Withdraw")

	fact, ok := op.Fact().(WithdrawFact)
	if !ok {
		return ctx, nil, e.Errorf("expected WithdrawFact, not %T", op.Fact())
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.sender), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("sender not found, %v; %w", fact.sender, err), nil
	}

	if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.sender), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account cannot be sender, %v; %w", fact.sender, err), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing; %w", err), nil
	}

	for i := range fact.items {
		cip := withdrawItemProcessorPool.Get()
		c, ok := cip.(*WithdrawItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected WithdrawItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.sender = fact.sender
		c.item = fact.items[i]

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess WithdrawItem; %w", err), nil
		}

		c.Close()
	}

	return ctx, nil, nil
}

func (opp *WithdrawProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(WithdrawFact)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected WithdrawFact, not %T", op.Fact()), nil
	}

	feeReceiveBalSts, required, err := opp.calculateItemsFee(op, getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to calculate fee: %v", err), nil
	}
	senderBalSts, err := currency.CheckEnoughBalance(fact.sender, required, getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check enough balance: %v", err), nil
	} else {
		opp.required = required
	}

	ns := make([]*WithdrawItemProcessor, len(fact.items))
	for i := range fact.items {
		cip := withdrawItemProcessorPool.Get()
		c, ok := cip.(*WithdrawItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected WithdrawItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.sender = fact.sender
		c.item = fact.items[i]

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess WithdrawItem: %v", err), nil
		}

		ns[i] = c
	}

	var stateMergeValues []base.StateMergeValue // nolint:prealloc
	for i := range ns {
		s, err := ns[i].Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process WithdrawItem: %v", err), nil
		}
		stateMergeValues = append(stateMergeValues, s...)

		ns[i].Close()
	}

	for cid := range senderBalSts {
		v, ok := senderBalSts[cid].Value().(statecurrency.BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", senderBalSts[cid].Value()), nil
		}

		var stateMergeValue base.StateMergeValue
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
				return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", feeReceiveBalSts[cid].Value()), nil
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

func (opp *WithdrawProcessor) Close() error {
	for i := range opp.ns {
		opp.ns[i].Close()
	}

	opp.required = nil
	withdrawProcessorPool.Put(opp)

	return nil
}

func (opp *WithdrawProcessor) calculateItemsFee(op base.Operation, getStateFunc base.GetStateFunc) (map[types.CurrencyID]base.State, map[types.CurrencyID][2]common.Big, error) {
	fact, ok := op.Fact().(WithdrawFact)
	if !ok {
		return nil, nil, errors.Errorf("expected WithdrawFact, not %T", op.Fact())
	}
	items := make([]currency.AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return currency.CalculateItemsFee(getStateFunc, items)
}
