package extension

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	types "github.com/ProtoconNet/mitum-currency/v3/operation/type"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"sync"

	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var createContractAccountsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateContractAccountsItemProcessor)
	},
}

var createContractAccountsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateContractAccountsProcessor)
	},
}

func (CreateContractAccounts) Process(
	_ context.Context, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type CreateContractAccountsItemProcessor struct {
	h      util.Hash
	sender mitumbase.Address
	item   CreateContractAccountsItem
	ns     mitumbase.StateMergeValue
	oas    mitumbase.StateMergeValue
	oac    base.Account
	nb     map[base.CurrencyID]mitumbase.StateMergeValue
}

func (opp *CreateContractAccountsItemProcessor) PreProcess(
	_ context.Context, _ mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
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

	st, err := state.NotExistsState(statecurrency.StateKeyAccount(target), "key of target account", getStateFunc)
	if err != nil {
		return err
	}
	opp.ns = state.NewStateMergeValue(st.Key(), st.Value())

	st, err = state.NotExistsState(extension.StateKeyContractAccount(target), "key of target contract account", getStateFunc)
	if err != nil {
		return err
	}
	opp.oas = state.NewStateMergeValue(st.Key(), st.Value())

	st, err = state.ExistsState(statecurrency.StateKeyAccount(opp.sender), "key of sender account", getStateFunc)
	if err != nil {
		return err
	}
	oac, err := statecurrency.LoadStateAccountValue(st)
	if err != nil {
		return err
	}
	opp.oac = oac

	nb := map[base.CurrencyID]mitumbase.StateMergeValue{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		switch _, found, err := getStateFunc(statecurrency.StateKeyBalance(target, am.Currency())); {
		case err != nil:
			return err
		case found:
			return isaac.ErrStopProcessingRetry.Errorf("target balance already exists, %q", target)
		default:
			nb[am.Currency()] = state.NewStateMergeValue(statecurrency.StateKeyBalance(target, am.Currency()), statecurrency.NewBalanceStateValue(base.NewZeroAmount(am.Currency())))
		}
	}
	opp.nb = nb

	return nil
}

func (opp *CreateContractAccountsItemProcessor) Process(
	_ context.Context, _ mitumbase.Operation, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, error) {
	e := util.StringErrorFunc("failed to preprocess for CreateContractAccountsItemProcessor")

	sts := make([]mitumbase.StateMergeValue, len(opp.item.Amounts())+2)

	var (
		nac base.Account
		err error
	)

	if opp.item.AddressType() == base.EthAddressHint.Type() {
		nac, err = base.NewEthAccountFromKeys(opp.item.Keys())
	} else {
		nac, err = base.NewAccountFromKeys(opp.item.Keys())
	}
	if err != nil {
		return nil, e(err, "")
	}

	ks, err := NewContractAccountKeys()
	if err != nil {
		return nil, e(err, "")
	}

	ncac, err := nac.SetKeys(ks)
	if err != nil {
		return nil, e(err, "")
	}
	sts[0] = state.NewStateMergeValue(opp.ns.Key(), statecurrency.NewAccountStateValue(ncac))

	cas := base.NewContractAccount(opp.oac.Address(), true)
	sts[1] = state.NewStateMergeValue(opp.oas.Key(), extension.NewContractAccountStateValue(cas))

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		v, ok := opp.nb[am.Currency()].Value().(statecurrency.BalanceStateValue)
		if !ok {
			return nil, errors.Errorf("expected BalanceStateValue, not %T", opp.nb[am.Currency()].Value())
		}
		stv := statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(am.Big())))
		sts[i+2] = state.NewStateMergeValue(opp.nb[am.Currency()].Key(), stv)
	}

	return sts, nil
}

func (opp *CreateContractAccountsItemProcessor) Close() {
	opp.h = nil
	opp.item = nil
	opp.ns = nil
	opp.nb = nil
	opp.sender = nil
	opp.oas = nil
	opp.oac = base.Account{}

	createContractAccountsItemProcessorPool.Put(opp)
}

type CreateContractAccountsProcessor struct {
	*mitumbase.BaseOperationProcessor
	ns       []*CreateContractAccountsItemProcessor
	required map[base.CurrencyID][2]base.Big // required[0] : amount + fee, required[1] : fee
}

func NewCreateContractAccountsProcessor() types.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new CreateContractAccountsProcessor")

		nopp := createContractAccountsProcessorPool.Get()
		opp, ok := nopp.(*CreateContractAccountsProcessor)
		if !ok {
			return nil, e(nil, "expected CreateContractAccountsProcessor, not %T", nopp)
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

func (opp *CreateContractAccountsProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	e := util.StringErrorFunc("failed to preprocess CreateContractAccounts")

	fact, ok := op.Fact().(CreateContractAccountsFact)
	if !ok {
		return ctx, nil, e(nil, "expected CreateContractAccountsFact, not %T", op.Fact())
	}

	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.sender), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.sender, err), nil
	}

	if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.sender), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account cannot be create-contract-account sender, %q: %w", fact.sender, err), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
	}

	return ctx, nil, nil
}

func (opp *CreateContractAccountsProcessor) Process( // nolint:dupl
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(CreateContractAccountsFact)
	if !ok {
		return nil, nil, mitumbase.NewBaseOperationProcessReasonError("expected CreateContractAccountsFact, not %T", op.Fact())
	}

	var (
		senderBalSts, feeReceiveBalSts map[base.CurrencyID]mitumbase.State
		required                       map[base.CurrencyID][2]base.Big
		err                            error
	)

	if feeReceiveBalSts, required, err = opp.calculateItemsFee(op, getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
	} else if senderBalSts, err = currency.CheckEnoughBalance(fact.sender, required, getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("not enough balance of sender %s : %w", fact.sender, err), nil
	} else {
		opp.required = required
	}

	ns := make([]*CreateContractAccountsItemProcessor, len(fact.items))
	for i := range fact.items {
		cip := createContractAccountsItemProcessorPool.Get()
		c, ok := cip.(*CreateContractAccountsItemProcessor)
		if !ok {
			return nil, mitumbase.NewBaseOperationProcessReasonError("expected CreateContractAccountsItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.item = fact.items[i]

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("fail to preprocess CreateContractAccountsItem: %w", err), nil
		}

		ns[i] = c
	}
	opp.ns = ns

	var stateMergeValues []mitumbase.StateMergeValue // nolint:prealloc
	for i := range ns {
		s, err := ns[i].Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to process CreateContractAccountsItem: %w", err), nil
		}
		stateMergeValues = append(stateMergeValues, s...)
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

func (opp *CreateContractAccountsProcessor) Close() error {
	for i := range opp.ns {
		opp.ns[i].Close()
	}

	opp.ns = nil
	opp.required = nil

	createContractAccountsProcessorPool.Put(opp)

	return nil
}

func (opp *CreateContractAccountsProcessor) calculateItemsFee(
	op mitumbase.Operation,
	getStateFunc mitumbase.GetStateFunc,
) (map[base.CurrencyID]mitumbase.State, map[base.CurrencyID][2]base.Big, error) {
	fact, ok := op.Fact().(CreateContractAccountsFact)
	if !ok {
		return nil, nil, errors.Errorf("expected CreateContractAccountsFact, not %T", op.Fact())
	}

	items := make([]currency.AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return currency.CalculateItemsFee(getStateFunc, items)
}
