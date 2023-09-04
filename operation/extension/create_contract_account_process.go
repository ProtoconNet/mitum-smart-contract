package extension

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	currencystate "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var createContractAccountItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateContractAccountItemProcessor)
	},
}

var createContractAccountProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateContractAccountProcessor)
	},
}

func (CreateContractAccount) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type CreateContractAccountItemProcessor struct {
	h      util.Hash
	sender base.Address
	item   CreateContractAccountItem
	ns     base.StateMergeValue
	oas    base.StateMergeValue
	oac    types.Account
	nb     map[types.CurrencyID]base.StateMergeValue
}

func (opp *CreateContractAccountItemProcessor) PreProcess(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) error {
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		policy, err := state.ExistsCurrencyPolicy(am.Currency(), getStateFunc)
		if err != nil {
			return err
		}

		if am.Big().Compare(policy.NewAccountMinBalance()) < 0 {
			return errors.Errorf("amount should be over minimum balance, %v < %v", am.Big(), policy.NewAccountMinBalance())
		}
	}

	target, err := opp.item.Address()
	if err != nil {
		return err
	}

	st, err := state.NotExistsState(currencystate.StateKeyAccount(target), "key of target account", getStateFunc)
	if err != nil {
		return err
	}
	opp.ns = state.NewStateMergeValue(st.Key(), st.Value())

	st, err = state.NotExistsState(extension.StateKeyContractAccount(target), "key of target contract account", getStateFunc)
	if err != nil {
		return err
	}
	opp.oas = state.NewStateMergeValue(st.Key(), st.Value())

	st, err = state.ExistsState(currencystate.StateKeyAccount(opp.sender), "key of sender account", getStateFunc)
	if err != nil {
		return err
	}
	oac, err := currencystate.LoadStateAccountValue(st)
	if err != nil {
		return err
	}
	opp.oac = oac

	nb := map[types.CurrencyID]base.StateMergeValue{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		switch _, found, err := getStateFunc(currencystate.StateKeyBalance(target, am.Currency())); {
		case err != nil:
			return err
		case found:
			return isaac.ErrStopProcessingRetry.Errorf("target balance already exists, %v", target)
		default:
			nb[am.Currency()] = state.NewStateMergeValue(currencystate.StateKeyBalance(target, am.Currency()), currencystate.NewBalanceStateValue(types.NewZeroAmount(am.Currency())))
		}
	}
	opp.nb = nb

	return nil
}

func (opp *CreateContractAccountItemProcessor) Process(
	_ context.Context, _ base.Operation, _ base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	e := util.StringError("failed to preprocess for CreateContractAccountItemProcessor")

	sts := make([]base.StateMergeValue, len(opp.item.Amounts())+2)

	var (
		nac types.Account
		err error
	)

	if opp.item.AddressType() == types.EthAddressHint.Type() {
		nac, err = types.NewEthAccountFromKeys(opp.item.Keys())
	} else {
		nac, err = types.NewAccountFromKeys(opp.item.Keys())
	}
	if err != nil {
		return nil, e.Wrap(err)
	}

	ks, err := types.NewContractAccountKeys()
	if err != nil {
		return nil, e.Wrap(err)
	}

	ncac, err := nac.SetKeys(ks)
	if err != nil {
		return nil, e.Wrap(err)
	}
	sts[0] = state.NewStateMergeValue(opp.ns.Key(), currencystate.NewAccountStateValue(ncac))

	cas := types.NewContractAccountStatus(opp.oac.Address())
	sts[1] = state.NewStateMergeValue(opp.oas.Key(), extension.NewContractAccountStateValue(cas))

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		v, ok := opp.nb[am.Currency()].Value().(currencystate.BalanceStateValue)
		if !ok {
			return nil, errors.Errorf("expected BalanceStateValue, not %T", opp.nb[am.Currency()].Value())
		}
		stv := currencystate.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(am.Big())))
		sts[i+2] = state.NewStateMergeValue(opp.nb[am.Currency()].Key(), stv)
	}

	return sts, nil
}

func (opp *CreateContractAccountItemProcessor) Close() {
	opp.h = nil
	opp.item = nil
	opp.ns = nil
	opp.nb = nil
	opp.sender = nil
	opp.oas = nil
	opp.oac = types.Account{}

	createContractAccountItemProcessorPool.Put(opp)
}

type CreateContractAccountProcessor struct {
	*base.BaseOperationProcessor
	ns       []*CreateContractAccountItemProcessor
	required map[types.CurrencyID][2]common.Big // required[0] : amount + fee, required[1] : fee
}

func NewCreateContractAccountProcessor() types.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new CreateContractAccountProcessor")

		nopp := createContractAccountProcessorPool.Get()
		opp, ok := nopp.(*CreateContractAccountProcessor)
		if !ok {
			return nil, e.Errorf("expected CreateContractAccountProcessor, not %T", nopp)
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

func (opp *CreateContractAccountProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("failed to preprocess CreateContractAccount")

	fact, ok := op.Fact().(CreateContractAccountFact)
	if !ok {
		return ctx, nil, e.Errorf("expected CreateContractAccountFact, not %T", op.Fact())
	}

	if err := state.CheckExistsState(currencystate.StateKeyAccount(fact.sender), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("sender not found, %v; %w", fact.sender, err), nil
	}

	if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.sender), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account cannot be create-contract-account sender, %v: %v", fact.sender, err), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing: %v", err), nil
	}

	for i := range fact.items {
		cip := createContractAccountItemProcessorPool.Get()
		c, ok := cip.(*CreateContractAccountItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected CreateContractAccountItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.item = fact.items[i]
		c.sender = fact.sender

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess CreateContractAccountItem; %w", err), nil
		}

		c.Close()
	}

	return ctx, nil, nil
}

func (opp *CreateContractAccountProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(CreateContractAccountFact)
	if !ok {
		return nil, nil, base.NewBaseOperationProcessReasonError("expected CreateContractAccountFact, not %T", op.Fact())
	}

	var (
		senderBalSts, feeReceiveBalSts map[types.CurrencyID]base.State
		required                       map[types.CurrencyID][2]common.Big
		err                            error
	)

	if feeReceiveBalSts, required, err = opp.calculateItemsFee(op, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to calculate fee: %v", err), nil
	} else if senderBalSts, err = currency.CheckEnoughBalance(fact.sender, required, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("not enough balance of sender %s : %v", fact.sender, err), nil
	} else {
		opp.required = required
	}

	ns := make([]*CreateContractAccountItemProcessor, len(fact.items))
	for i := range fact.items {
		cip := createContractAccountItemProcessorPool.Get()
		c, ok := cip.(*CreateContractAccountItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected CreateContractAccountItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.item = fact.items[i]
		c.sender = fact.sender

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess CreateContractAccountItem: %v", err), nil
		}

		ns[i] = c
	}
	opp.ns = ns

	var stateMergeValues []base.StateMergeValue // nolint:prealloc
	for i := range ns {
		s, err := ns[i].Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process CreateContractAccountItem: %v", err), nil
		}
		stateMergeValues = append(stateMergeValues, s...)
	}

	for cid := range senderBalSts {
		v, ok := senderBalSts[cid].Value().(currencystate.BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", senderBalSts[cid].Value()), nil
		}

		var stateMergeValue base.StateMergeValue
		if senderBalSts[cid].Key() == feeReceiveBalSts[cid].Key() {
			stateMergeValue = state.NewStateMergeValue(
				senderBalSts[cid].Key(),
				currencystate.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[cid][0]).Add(opp.required[cid][1]))),
			)
		} else {
			stateMergeValue = state.NewStateMergeValue(
				senderBalSts[cid].Key(),
				currencystate.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[cid][0]))),
			)
			r, ok := feeReceiveBalSts[cid].Value().(currencystate.BalanceStateValue)
			if !ok {
				return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", feeReceiveBalSts[cid].Value()), nil
			}
			stateMergeValues = append(
				stateMergeValues,
				state.NewStateMergeValue(
					feeReceiveBalSts[cid].Key(),
					currencystate.NewBalanceStateValue(r.Amount.WithBig(r.Amount.Big().Add(opp.required[cid][1]))),
				),
			)
		}
		stateMergeValues = append(stateMergeValues, stateMergeValue)
	}

	return stateMergeValues, nil, nil
}

func (opp *CreateContractAccountProcessor) Close() error {
	for i := range opp.ns {
		opp.ns[i].Close()
	}

	opp.ns = nil
	opp.required = nil

	createContractAccountProcessorPool.Put(opp)

	return nil
}

func (opp *CreateContractAccountProcessor) calculateItemsFee(
	op base.Operation,
	getStateFunc base.GetStateFunc,
) (map[types.CurrencyID]base.State, map[types.CurrencyID][2]common.Big, error) {
	fact, ok := op.Fact().(CreateContractAccountFact)
	if !ok {
		return nil, nil, errors.Errorf("expected CreateContractAccountFact, not %T", op.Fact())
	}

	items := make([]currency.AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return currency.CalculateItemsFee(getStateFunc, items)
}
