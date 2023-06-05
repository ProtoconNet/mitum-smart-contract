package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/base"
	types "github.com/ProtoconNet/mitum-currency/v3/operation/type"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"sync"

	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var keyUpdaterProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(KeyUpdaterProcessor)
	},
}

func (KeyUpdater) Process(
	_ context.Context, _ mitumbase.GetStateFunc,
) ([]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type KeyUpdaterProcessor struct {
	*mitumbase.BaseOperationProcessor
	sa  mitumbase.StateMergeValue
	sb  mitumbase.StateMergeValue
	fee base.Big
	// collectFee func(AddFee) error
}

func NewKeyUpdaterProcessor(
// collectFee func(*OperationProcessor, AddFee) error,
) types.GetNewProcessor {
	return func(
		height mitumbase.Height,
		getStateFunc mitumbase.GetStateFunc,
		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	) (mitumbase.OperationProcessor, error) {
		e := util.StringErrorFunc("failed to create new KeyUpdaterProcessor")

		nopp := keyUpdaterProcessorPool.Get()
		opp, ok := nopp.(*KeyUpdaterProcessor)
		if !ok {
			return nil, errors.Errorf("expected KeyUpdaterProcessor, not %T", nopp)
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

func (opp *KeyUpdaterProcessor) PreProcess(
	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
) (context.Context, mitumbase.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(KeyUpdaterFact)
	if !ok {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("expected KeyUpdaterFact, not %T", op.Fact()), nil
	}

	if err := state.CheckFactSignsByState(fact.target, op.Signs(), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing :  %w", err), nil
	}

	if st, err := state.ExistsState(currency.StateKeyAccount(fact.target), "target keys", getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("failed to check existence of target %v : %w", fact.target, err), nil
	} else if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.Target()), getStateFunc); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account not allowed for key updater, %q: %w", fact.Target(), err), nil
	} else if ks, err := currency.StateKeysValue(st); err != nil {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("failed to get state value of keys %v : %w", fact.keys.Hash(), err), nil
	} else if ks.Equal(fact.Keys()) {
		return ctx, mitumbase.NewBaseOperationProcessReasonError("same Keys as existing %v : %w", fact.keys.Hash(), err), nil
	}

	return ctx, nil, nil
}

func (opp *KeyUpdaterProcessor) Process( // nolint:dupl
	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
) {
	e := util.StringErrorFunc("failed to process KeyUpdater")

	fact, ok := op.Fact().(KeyUpdaterFact)
	if !ok {
		return nil, nil, e(nil, "expected KeyUpdaterFact, not %T", op.Fact())
	}

	var tgAccSt mitumbase.State
	var err error
	if tgAccSt, err = state.ExistsState(currency.StateKeyAccount(fact.target), "target keys", getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check existence of target %v : %w", fact.target, err), nil
	}

	var fee base.Big
	var policy base.CurrencyPolicy
	if policy, err = state.ExistsCurrencyPolicy(fact.currency, getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check existence of currency %v : %w", fact.currency, err), nil
	} else if fee, err = policy.Feeer().Fee(base.ZeroBig); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check fee of currency %v : %w", fact.currency, err), nil
	}

	var tgBalSt mitumbase.State
	if tgBalSt, err = state.ExistsState(currency.StateKeyBalance(fact.target, fact.currency), "balance of target", getStateFunc); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check existence of targe balance %v : %w", fact.target, err), nil
	} else if b, err := currency.StateBalanceValue(tgBalSt); err != nil {
		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check existence of target balance %v,%v : %w", fact.currency, fact.target, err), nil
	} else if b.Big().Compare(fee) < 0 {
		return nil, mitumbase.NewBaseOperationProcessReasonError("insufficient balance with fee %v,%v", fact.currency, fact.target), nil
	}

	var stmvs []mitumbase.StateMergeValue // nolint:prealloc
	v, ok := tgBalSt.Value().(currency.BalanceStateValue)
	if !ok {
		return nil, mitumbase.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", tgBalSt.Value()), nil
	}

	tgAmount := v.Amount.WithBig(v.Amount.Big().Sub(fee))

	// stv := NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(opp.required[i][0]).Sub(opp.required[i][1])))
	if policy.Feeer().Receiver() != nil {
		if err := state.CheckExistsState(currency.StateKeyAccount(policy.Feeer().Receiver()), getStateFunc); err != nil {
			return nil, nil, err
		} else if feeRcvrSt, found, err := getStateFunc(currency.StateKeyBalance(policy.Feeer().Receiver(), fact.currency)); err != nil {
			return nil, nil, err
		} else if !found {
			return nil, nil, errors.Errorf("feeer receiver %s not found", policy.Feeer().Receiver())
		} else if feeRcvrSt.Key() == tgBalSt.Key() {
			tgAmount = tgAmount.WithBig(tgAmount.Big().Add(fee))
		} else {
			r, ok := feeRcvrSt.Value().(currency.BalanceStateValue)
			if !ok {
				return nil, nil, errors.Errorf("invalid BalanceState value found, %T", feeRcvrSt.Value())
			}
			stmvs = append(stmvs, state.NewStateMergeValue(feeRcvrSt.Key(), currency.NewBalanceStateValue(r.Amount.WithBig(r.Amount.Big().Add(fee)))))
		}
	}
	stmv := currency.NewBalanceStateValue(tgAmount)
	stmvs = append(stmvs, state.NewStateMergeValue(tgBalSt.Key(), stmv))

	ac, err := currency.LoadStateAccountValue(tgAccSt)
	if err != nil {
		return nil, nil, err
	}
	uac, err := ac.SetKeys(fact.keys)
	if err != nil {
		return nil, nil, err
	}
	stmvs = append(stmvs, state.NewStateMergeValue(tgAccSt.Key(), currency.NewAccountStateValue(uac)))

	return stmvs, nil, nil
}

func (opp *KeyUpdaterProcessor) Close() error {
	keyUpdaterProcessorPool.Put(opp)

	return nil
}
