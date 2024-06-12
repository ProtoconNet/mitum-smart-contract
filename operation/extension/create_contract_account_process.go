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
	e := util.StringError("preprocess CreateContractAccountItemProcessor")

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		policy, err := state.ExistsCurrencyPolicy(am.Currency(), getStateFunc)
		if err != nil {
			return e.Wrap(err)
		}

		if am.Big().Compare(policy.MinBalance()) < 0 {
			return e.Wrap(common.ErrValOOR.Wrap(errors.Errorf("amount under new account minimum balance, %v < %v", am.Big(), policy.MinBalance())))

		}
	}

	target, err := opp.item.Address()
	if err != nil {
		return e.Wrap(err)
	}

	ast, cst, aErr, cErr := state.ExistsCAccount(target, "target", false, false, getStateFunc)
	if aErr != nil {
		return e.Wrap(aErr)
	} else if cErr != nil {
		return e.Wrap(cErr)
	}

	opp.ns = state.NewStateMergeValue(ast.Key(), ast.Value())
	opp.oas = state.NewStateMergeValue(cst.Key(), cst.Value())

	aSt, aErr := state.ExistsAccount(opp.sender, "sender", true, getStateFunc)
	if aErr != nil {
		return e.Wrap(aErr)
	}

	oac, err := currencystate.LoadAccountStateValue(aSt)
	if err != nil {
		return e.Wrap(err)
	}

	opp.oac = *oac

	nb := map[types.CurrencyID]base.StateMergeValue{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		switch _, found, err := getStateFunc(currencystate.BalanceStateKey(target, am.Currency())); {
		case err != nil:
			return e.Wrap(err)
		case found:
			return e.Wrap(common.ErrAccountE.Wrap(errors.Errorf("target account balance already exists, %v", target)))

		default:
			nb[am.Currency()] = common.NewBaseStateMergeValue(
				currencystate.BalanceStateKey(target, am.Currency()),
				currencystate.NewAddBalanceStateValue(types.NewZeroAmount(am.Currency())),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currencystate.NewBalanceStateValueMerger(
						height,
						currencystate.BalanceStateKey(target, am.Currency()), am.Currency(), st)
				},
			)
		}
	}
	opp.nb = nb

	return nil
}

func (opp *CreateContractAccountItemProcessor) Process(
	_ context.Context, _ base.Operation, _ base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	e := util.StringError("process for CreateContractAccountItemProcessor")

	sts := make([]base.StateMergeValue, len(opp.item.Amounts())+2)

	nac, err := types.NewAccountFromKeys(opp.item.Keys())

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

	cas := types.NewContractAccountStatus(opp.oac.Address(), nil)
	sts[1] = state.NewStateMergeValue(opp.oas.Key(), extension.NewContractAccountStateValue(cas))

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		v, ok := opp.nb[am.Currency()].Value().(currencystate.AddBalanceStateValue)
		if !ok {
			return nil, errors.Errorf("expected AddBalanceStateValue, not %T", opp.nb[am.Currency()].Value())
		}
		sts[i+2] = common.NewBaseStateMergeValue(
			opp.nb[am.Currency()].Key(),
			currencystate.NewAddBalanceStateValue(v.Amount.WithBig(am.Big())),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currencystate.NewBalanceStateValueMerger(height, opp.nb[am.Currency()].Key(), am.Currency(), st)
			},
		)

		//stv := currencystate.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(am.Big())))
		//sts[i+2] = state.NewStateMergeValue(opp.nb[am.Currency()].Key(), stv)
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
		e := util.StringError("create new CreateContractAccountProcessor")

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
	fact, ok := op.Fact().(CreateContractAccountFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected CreateContractAccountFact, not %T", op.Fact())), nil
	}

	aSt, _, aErr, cErr := state.ExistsCAccount(fact.Sender(), "sender", true, false, getStateFunc)
	if aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMCAccountNA).Errorf("%v", cErr)), nil
	}

	ac, err := currencystate.LoadAccountStateValue(aSt)
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMStateValInvalid).
				Errorf("%v: sender account, %v", err, fact.Sender())), nil
	}

	if err := state.CheckFactSignsByState(fact.Sender(), op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMSignInvalid).
				Errorf("%v", err)), nil
	}

	for i := range fact.items {
		cip := createContractAccountItemProcessorPool.Get()
		c, ok := cip.(*CreateContractAccountItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Wrap(common.ErrMTypeMismatch).
					Errorf("expected CreateContractAccountItemProcessor, not %T", cip)), nil
		}

		c.h = op.Hash()
		c.item = fact.items[i]
		c.sender = fact.Sender()
		c.oac = *ac

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError(
				common.ErrMPreProcess.
					Errorf("%v", err)), nil
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
		senderBalSts, feeReceiverBalSts map[types.CurrencyID]base.State
		required                        map[types.CurrencyID][2]common.Big
		err                             error
	)

	if feeReceiverBalSts, required, err = opp.calculateItemsFee(op, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("calculate fee; %w", err), nil
	} else if senderBalSts, err = currency.CheckEnoughBalance(fact.Sender(), required, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("not enough balance of sender %s; %v", fact.Sender(), err), nil
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
		c.sender = fact.Sender()

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
			return nil, base.NewBaseOperationProcessReasonError("process CreateContractAccountItem: %v", err), nil
		}
		stateMergeValues = append(stateMergeValues, s...)
	}

	for cid := range senderBalSts {
		v, ok := senderBalSts[cid].Value().(currencystate.BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", senderBalSts[cid].Value()), nil
		}
		_, feeReceiverFound := feeReceiverBalSts[cid]

		var stateMergeValue base.StateMergeValue
		if feeReceiverFound && (senderBalSts[cid].Key() == feeReceiverBalSts[cid].Key()) {
			stateMergeValue = common.NewBaseStateMergeValue(
				senderBalSts[cid].Key(),
				currencystate.NewDeductBalanceStateValue(v.Amount.WithBig(opp.required[cid][0].Sub(opp.required[cid][1]))),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currencystate.NewBalanceStateValueMerger(height, senderBalSts[cid].Key(), cid, st)
				},
			)
		} else {
			stateMergeValue = common.NewBaseStateMergeValue(
				senderBalSts[cid].Key(),
				currencystate.NewDeductBalanceStateValue(v.Amount.WithBig(opp.required[cid][0])),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currencystate.NewBalanceStateValueMerger(height, senderBalSts[cid].Key(), cid, st)
				},
			)
			if feeReceiverFound {
				r, ok := feeReceiverBalSts[cid].Value().(currencystate.BalanceStateValue)
				if !ok {
					return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", feeReceiverBalSts[cid].Value()), nil
				}
				stateMergeValues = append(
					stateMergeValues,
					common.NewBaseStateMergeValue(
						feeReceiverBalSts[cid].Key(),
						currencystate.NewAddBalanceStateValue(r.Amount.WithBig(opp.required[cid][1])),
						func(height base.Height, st base.State) base.StateValueMerger {
							return currencystate.NewBalanceStateValueMerger(height, feeReceiverBalSts[cid].Key(), cid, st)
						},
					),
				)
			}
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
