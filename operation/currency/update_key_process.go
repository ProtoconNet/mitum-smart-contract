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

	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var updateKeyProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(UpdateKeyProcessor)
	},
}

func (UpdateKey) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type UpdateKeyProcessor struct {
	*base.BaseOperationProcessor
}

func NewUpdateKeyProcessor(
// collectFee func(*OperationProcessor, AddFee) error,
) types.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new UpdateKeyProcessor")

		nopp := updateKeyProcessorPool.Get()
		opp, ok := nopp.(*UpdateKeyProcessor)
		if !ok {
			return nil, errors.Errorf("expected UpdateKeyProcessor, not %T", nopp)
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

func (opp *UpdateKeyProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(UpdateKeyFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError("expected UpdateKeyFact, not %T", op.Fact()), nil
	}

	if err := state.CheckFactSignsByState(fact.target, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing; %w", err), nil
	}

	if st, err := state.ExistsState(currency.StateKeyAccount(fact.target), "target keys", getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("failed to check existence of target %v; %w", fact.target, err), nil
	} else if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.Target()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account not allowed for key updater, %v; %w", fact.Target(), err), nil
	} else if ks, err := currency.StateKeysValue(st); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("failed to get state value of keys %q ; %w", fact.keys.Hash(), err), nil
	} else if ks.Equal(fact.Keys()) {
		return ctx, base.NewBaseOperationProcessReasonError("same Keys as existing %q; %w", fact.keys.Hash(), err), nil
	}

	return ctx, nil, nil
}

func (opp *UpdateKeyProcessor) Process( // nolint:dupl
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("failed to process UpdateKey")

	fact, ok := op.Fact().(UpdateKeyFact)
	if !ok {
		return nil, nil, e.Errorf("expected UpdateKeyFact, not %T", op.Fact())
	}

	var tgAccSt base.State
	var err error
	if tgAccSt, err = state.ExistsState(currency.StateKeyAccount(fact.target), "target keys", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of target %v ; %w", fact.target, err), nil
	}

	var fee common.Big
	var policy types.CurrencyPolicy
	if policy, err = state.ExistsCurrencyPolicy(fact.currency, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of currency %v; %w", fact.currency, err), nil
	} else if fee, err = policy.Feeer().Fee(common.ZeroBig); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check fee of currency %v; %w", fact.currency, err), nil
	}

	var tgBalSt base.State
	if tgBalSt, err = state.ExistsState(currency.StateKeyBalance(fact.target, fact.currency), "balance of target", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of target balance %v ; %w", fact.target, err), nil
	} else if b, err := currency.StateBalanceValue(tgBalSt); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to check existence of target balance %v, %v ; %w", fact.currency, fact.target, err), nil
	} else if b.Big().Compare(fee) < 0 {
		return nil, base.NewBaseOperationProcessReasonError("insufficient balance with fee %v ,%v", fact.currency, fact.target), nil
	}

	var stmvs []base.StateMergeValue // nolint:prealloc
	v, ok := tgBalSt.Value().(currency.BalanceStateValue)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", tgBalSt.Value()), nil
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

func (opp *UpdateKeyProcessor) Close() error {
	updateKeyProcessorPool.Put(opp)

	return nil
}
