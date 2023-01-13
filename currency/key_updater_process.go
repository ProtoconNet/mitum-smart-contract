package currency

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
)

var keyUpdaterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(KeyUpdaterProcessor)
	},
}

func (KeyUpdater) Process(
	ctx context.Context, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type KeyUpdaterProcessor struct {
	*base.BaseOperationProcessor
	sa  base.StateMergeValue
	sb  base.StateMergeValue
	fee Big
	// collectFee func(AddFee) error
}

func NewKeyUpdaterProcessor(
// collectFee func(*OperationProcessor, AddFee) error,
) GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new KeyUpdaterProcessor")

		nopp := keyUpdaterProcessorPool.Get()
		opp, ok := nopp.(*KeyUpdaterProcessor)
		if !ok {
			return nil, errors.Errorf("expected KeyUpdaterProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e(err, "")
		}

		opp.BaseOperationProcessor = b
		opp.sb = nil
		opp.sa = nil
		opp.fee = ZeroBig
		// opp.collectFee = collectFee

		return opp, nil
	}
}

func (opp *KeyUpdaterProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(KeyUpdaterFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError("expected KeyUpdaterFact, not %T", op.Fact()), nil
	}

	if err := checkFactSignsByState(fact.target, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing :  %w", err), nil
	}

	if st, err := existsState(StateKeyAccount(fact.target), "target keys", getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("failed to check existence of target %v : %w", fact.target, err), nil
	} else if ks, err := StateKeysValue(st); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("failed to get state value of keys %v : %w", fact.keys.Hash(), err), nil
	} else if ks.Equal(fact.Keys()) {
		return ctx, base.NewBaseOperationProcessReasonError("same Keys as existing %v : %w", fact.keys.Hash(), err), nil
	}

	return ctx, nil, nil
}

func (opp *KeyUpdaterProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(KeyUpdaterFact)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected KeyUpdaterFact, not %T", op.Fact()), nil
	}

	if st, err := existsState(StateKeyAccount(fact.target), "target keys", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of target %v : %w", fact.target, err), nil
	} else {
		opp.sa = NewAccountStateMergeValue(st.Key(), st.Value())
	}

	var fee Big
	if policy, err := existsCurrencyPolicy(fact.currency, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of currency %v : %w", fact.currency, err), nil
	} else if k, err := policy.Feeer().Fee(ZeroBig); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check fee of currency %v : %w", fact.currency, err), nil
	} else {
		fee = k
	}

	var bst base.State
	if st, err := existsState(StateKeyBalance(fact.target, fact.currency), "balance of target", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of targe balance %v : %w", fact.target, err), nil
	} else {
		bst = st
		opp.sb = NewBalanceStateMergeValue(st.Key(), st.Value())
	}

	switch b, err := StateBalanceValue(bst); {
	case err != nil:
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of target balance %v,%v : %w", fact.currency, fact.target, err), nil
	case b.Big().Compare(fee) < 0:
		return nil, base.NewBaseOperationProcessReasonError("insufficient balance with fee %v,%v", fact.currency, fact.target), nil
	default:
		opp.fee = fee
	}

	var sts []base.StateMergeValue // nolint:prealloc

	v, ok := opp.sb.Value().(BalanceStateValue)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", opp.sb.Value()), nil
	}
	stv := NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.fee)))
	// stv := NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[i][0]).Sub(opp.required[i][1])))
	sts = append(sts, NewBalanceStateMergeValue(opp.sb.Key(), stv))

	if a, err := NewAccountFromKeys(fact.keys); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to create new account from keys"), nil
	} else {
		v := NewAccountStateValue(a)
		sts = append(sts, NewAccountStateMergeValue(opp.sa.Key(), v))
	}

	return sts, nil, nil
}

func (opp *KeyUpdaterProcessor) Close() error {
	opp.sa = nil
	opp.sb = nil
	opp.fee = ZeroBig

	keyUpdaterProcessorPool.Put(opp)

	return nil
}
