package currency

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/util"
)

var createAccountsItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateAccountsItemProcessor)
	},
}

var createAccountsProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CreateAccountsProcessor)
	},
}

func (CreateAccounts) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type CreateAccountsItemProcessor struct {
	h    util.Hash
	item CreateAccountsItem
	ns   base.StateMergeValue
	nb   map[CurrencyID]base.StateMergeValue
}

func (opp *CreateAccountsItemProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) error {
	e := util.StringErrorFunc("failed to preprocess for CreateAccountsItemProcessor")

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]

		policy, err := existsCurrencyPolicy(am.cid, getStateFunc)
		if err != nil {
			return err
		}

		if am.Big().Compare(policy.NewAccountMinBalance()) < 0 {
			return base.NewBaseOperationProcessReasonError(
				"amount should be over minimum balance, %v < %v", am.Big(), policy.NewAccountMinBalance())
		}
	}

	target, err := opp.item.Address()
	if err != nil {
		return e(err, "")
	}

	st, err := notExistsState(StateKeyAccount(target), "keys of target", getStateFunc)
	if err != nil {
		return err
	}
	opp.ns = NewAccountStateMergeValue(st.Key(), st.Value())

	nb := map[CurrencyID]base.StateMergeValue{}
	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		switch _, found, err := getStateFunc(StateKeyBalance(target, am.Currency())); {
		case err != nil:
			return e(err, "")
		case found:
			return e(isaac.ErrStopProcessingRetry.Errorf("target balance already exists"), "")
		default:
			nb[am.Currency()] = NewBalanceStateMergeValue(StateKeyBalance(target, am.Currency()), NewBalanceStateValue(NewZeroAmount(am.Currency())))
		}
	}
	opp.nb = nb

	return nil
}

func (opp *CreateAccountsItemProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	e := util.StringErrorFunc("failed to preprocess for CreateAccountsItemProcessor")

	nac, err := NewAccountFromKeys(opp.item.Keys())
	if err != nil {
		return nil, e(err, "")
	}

	sts := make([]base.StateMergeValue, len(opp.item.Amounts())+1)
	sts[0] = NewAccountStateMergeValue(opp.ns.Key(), NewAccountStateValue(nac))

	for i := range opp.item.Amounts() {
		am := opp.item.Amounts()[i]
		v, ok := opp.nb[am.Currency()].Value().(BalanceStateValue)
		if !ok {
			return nil, e(errors.Errorf("not BalanceStateValue, %T", opp.nb[am.Currency()].Value()), "")
		}
		stv := NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(am.big)))
		sts[i+1] = NewBalanceStateMergeValue(opp.nb[am.Currency()].Key(), stv)
	}

	return sts, nil
}

func (opp *CreateAccountsItemProcessor) Close() error {
	opp.h = nil
	opp.item = nil
	opp.ns = nil
	opp.nb = nil

	createAccountsItemProcessorPool.Put(opp)

	return nil
}

type CreateAccountsProcessor struct {
	*base.BaseOperationProcessor
	ns       []*CreateAccountsItemProcessor
	required map[CurrencyID][2]Big // required[0] : amount + fee, required[1] : fee
}

func NewCreateAccountsProcessor() GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new CreateAccountsProcessor")

		nopp := createAccountsProcessorPool.Get()
		opp, ok := nopp.(*CreateAccountsProcessor)
		if !ok {
			return nil, errors.Errorf("expected CreateAccountsProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b
		opp.ns = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *CreateAccountsProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(CreateAccountsFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError("expected CreateAccountsFact, not %T", op.Fact()), nil
	}

	if err := checkExistsState(StateKeyAccount(fact.sender), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("failed to check existence of sender %v : %w", fact.sender, err), nil
	}

	if err := checkFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing :  %w", err), nil
	}

	return ctx, nil, nil
}

func (opp *CreateAccountsProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(CreateAccountsFact)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected CreateAccountsFact, not %T", op.Fact()), nil
	}

	var (
		sendrBalSts, feeRecvrBalSts map[CurrencyID]base.State
		required                    map[CurrencyID][2]Big
		err                         error
	)

	if feeRecvrBalSts, required, err = opp.calculateItemsFee(op, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
	} else if sendrBalSts, err = CheckEnoughBalance(fact.sender, required, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("not enough balance of sender %s : %w", fact.sender, err), nil
	} else {
		opp.required = required
	}

	ns := make([]*CreateAccountsItemProcessor, len(fact.items))
	for i := range fact.items {
		cip := createAccountsItemProcessorPool.Get()
		c, ok := cip.(*CreateAccountsItemProcessor)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected CreateAccountsItemProcessor, not %T", cip), nil
		}

		c.h = op.Hash()
		c.item = fact.items[i]

		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
			return nil, base.NewBaseOperationProcessReasonError("fail to preprocess CreateAccountsItem: %w", err), nil
		}

		ns[i] = c
	}
	opp.ns = ns

	var stmvs []base.StateMergeValue // nolint:prealloc
	for i := range opp.ns {
		s, err := opp.ns[i].Process(ctx, op, getStateFunc)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("failed to process CreateAccountsItem: %w", err), nil
		}
		stmvs = append(stmvs, s...)
	}

	for cid := range sendrBalSts {
		v, ok := sendrBalSts[cid].Value().(BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", sendrBalSts[cid].Value()), nil
		}

		var stmv base.StateMergeValue
		if sendrBalSts[cid].Key() == feeRecvrBalSts[cid].Key() {
			stmv = NewBalanceStateMergeValue(
				sendrBalSts[cid].Key(),
				NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[cid][0]).Add(opp.required[cid][1]))),
			)
		} else {
			stmv = NewBalanceStateMergeValue(
				sendrBalSts[cid].Key(),
				NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[cid][0]))),
			)
			r, ok := feeRecvrBalSts[cid].Value().(BalanceStateValue)
			if !ok {
				return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", feeRecvrBalSts[cid].Value()), nil
			}
			stmvs = append(
				stmvs,
				NewBalanceStateMergeValue(
					feeRecvrBalSts[cid].Key(),
					NewBalanceStateValue(r.Amount.WithBig(r.Amount.big.Add(opp.required[cid][1]))),
				),
			)
		}
		stmvs = append(stmvs, stmv)
	}

	return stmvs, nil, nil
}

func (opp *CreateAccountsProcessor) Close() error {
	for i := range opp.ns {
		_ = opp.ns[i].Close()
	}

	opp.ns = nil
	opp.required = nil

	createAccountsProcessorPool.Put(opp)

	return nil
}

func (opp *CreateAccountsProcessor) calculateItemsFee(op base.Operation, getStateFunc base.GetStateFunc) (map[CurrencyID]base.State, map[CurrencyID][2]Big, error) {
	fact, ok := op.Fact().(CreateAccountsFact)
	if !ok {
		return nil, nil, errors.Errorf("expected CreateAccountsFact, not %T", op.Fact())
	}

	items := make([]AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return CalculateItemsFee(getStateFunc, items)
}

func CalculateItemsFee(getStateFunc base.GetStateFunc, items []AmountsItem) (map[CurrencyID]base.State, map[CurrencyID][2]Big, error) {
	feeRecvrSts := map[CurrencyID]base.State{}
	required := map[CurrencyID][2]Big{}

	for i := range items {
		it := items[i]

		for j := range it.Amounts() {
			am := it.Amounts()[j]

			rq := [2]Big{ZeroBig, ZeroBig}
			if k, found := required[am.Currency()]; found {
				rq = k
			}

			policy, err := existsCurrencyPolicy(am.cid, getStateFunc)
			if err != nil {
				return nil, nil, err
			}

			var k Big
			switch k, err = policy.Feeer().Fee(am.Big()); {
			case err != nil:
				return nil, nil, err
			case !k.OverZero():
				required[am.Currency()] = [2]Big{rq[0].Add(am.Big()), rq[1]}
			default:
				required[am.Currency()] = [2]Big{rq[0].Add(am.Big()).Add(k), rq[1].Add(k)}
			}

			if policy.feeer.Receiver() == nil {
				continue
			}

			if err := checkExistsState(StateKeyAccount(policy.feeer.Receiver()), getStateFunc); err != nil {
				return nil, nil, err
			} else if st, _, err := getStateFunc(StateKeyBalance(policy.feeer.Receiver(), am.Currency())); err != nil {
				return nil, nil, err
			} else {
				feeRecvrSts[am.Currency()] = st
			}
		}
	}

	return feeRecvrSts, required, nil
}

func CheckEnoughBalance(
	holder base.Address,
	required map[CurrencyID][2]Big,
	getStateFunc base.GetStateFunc,
) (map[CurrencyID]base.State, error) {
	sbSts := map[CurrencyID]base.State{}

	for cid := range required {
		rq := required[cid]

		st, err := existsState(StateKeyBalance(holder, cid), "currency of holder", getStateFunc)
		if err != nil {
			return nil, err
		}

		am, err := StateBalanceValue(st)
		if err != nil {
			return nil, base.NewBaseOperationProcessReasonError("insufficient balance of sender: %w", err)
		}

		if am.Big().Compare(rq[0]) < 0 {
			return nil, base.NewBaseOperationProcessReasonError(
				"insufficient balance of sender, %s; %d !> %d", holder.String(), am.Big(), rq[0])
		}
		sbSts[cid] = st
	}

	return sbSts, nil
}
