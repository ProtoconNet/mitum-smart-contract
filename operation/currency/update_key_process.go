package currency

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
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

func NewUpdateKeyProcessor() types.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("create new UpdateKeyProcessor")

		nopp := updateKeyProcessorPool.Get()
		opp, ok := nopp.(*UpdateKeyProcessor)
		if !ok {
			return nil, errors.Errorf("expected %T, not %T", &UpdateKeyProcessor{}, nopp)
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
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMTypeMismatch).Errorf("expected %T, not %T", UpdateKeyFact{}, op.Fact())), nil
	}

	_, err := state.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err),
		), nil
	}

	if aState, _, aErr, cErr := state.ExistsCAccount(fact.Target(), "target", true, false, getStateFunc); aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).Errorf("%v", cErr)), nil
	} else if ac, err := currency.LoadStateAccountValue(aState); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMStateValInvalid).Errorf("%v: target %v", err, fact.Target())), nil
	} else if _, ok := ac.Keys().(types.NilAccountKeys); ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).Errorf("target %v must be multi-sig account", fact.Target())), nil
	} else if ac.Keys().Equal(fact.Keys()) {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMValueInvalid).Errorf("target keys is same with keys to update, keys hash %v", fact.keys.Hash())), nil
	}

	if err := state.CheckFactSignsByState(fact.Target(), op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMSignInvalid).Errorf("%v", err)), nil
	}

	return ctx, nil, nil
}

func (opp *UpdateKeyProcessor) Process( // nolint:dupl
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("process UpdateKey")

	fact, ok := op.Fact().(UpdateKeyFact)
	if !ok {
		return nil, nil, e.Errorf("expected %T, not %T", UpdateKeyFact{}, op.Fact())
	}

	var tgAccSt base.State
	var err error
	if tgAccSt, err = state.ExistsState(currency.StateKeyAccount(fact.Target()), "target keys", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("target account not found, %v; %w", fact.Target(), err), nil
	}

	var fee common.Big
	policy, err := state.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check existence of currency %v; %w", fact.Currency(), err), nil
	} else if fee, err = policy.Feeer().Fee(common.ZeroBig); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check fee of currency %v; %w", fact.Currency(), err), nil
	}

	var tgBalSt base.State
	if tgBalSt, err = state.ExistsState(currency.StateKeyBalance(fact.Target(), fact.Currency()), "balance of target", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("target account balance not found, %v; %w", fact.Target(), err), nil
	} else if b, err := currency.StateBalanceValue(tgBalSt); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("failed to get target balance value, %v, %v; %w", fact.Currency(), fact.Target(), err), nil
	} else if b.Big().Compare(fee) < 0 {
		return nil, base.NewBaseOperationProcessReasonError("insufficient balance with fee %v ,%v", fact.Currency(), fact.Target()), nil
	}

	var stmvs []base.StateMergeValue // nolint:prealloc
	v, ok := tgBalSt.Value().(currency.BalanceStateValue)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected %T, not %T", currency.BalanceStateValue{}, tgBalSt.Value()), nil
	}

	if policy.Feeer().Receiver() != nil {
		if err := state.CheckExistsState(currency.StateKeyAccount(policy.Feeer().Receiver()), getStateFunc); err != nil {
			return nil, nil, errors.Errorf("feeer receiver %s not found", policy.Feeer().Receiver())
		} else if feeRcvrSt, found, err := getStateFunc(currency.StateKeyBalance(policy.Feeer().Receiver(), fact.Currency())); err != nil {
			return nil, nil, errors.Errorf("feeer receiver %s balance of %s not found", policy.Feeer().Receiver(), fact.Currency())
		} else if !found {
			return nil, nil, errors.Errorf("feeer receiver %s balance of %s not found", policy.Feeer().Receiver(), fact.Currency())
		} else if feeRcvrSt.Key() != tgBalSt.Key() {
			r, ok := feeRcvrSt.Value().(currency.BalanceStateValue)
			if !ok {
				return nil, nil, errors.Errorf("expected %T, not %T", currency.BalanceStateValue{}, feeRcvrSt.Value())
			}
			stmvs = append(stmvs, common.NewBaseStateMergeValue(
				feeRcvrSt.Key(),
				currency.NewAddBalanceStateValue(r.Amount.WithBig(fee)),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currency.NewBalanceStateValueMerger(height, feeRcvrSt.Key(), fact.Currency(), st)
				},
			))

			stmvs = append(stmvs, common.NewBaseStateMergeValue(
				tgBalSt.Key(),
				currency.NewDeductBalanceStateValue(v.Amount.WithBig(fee)),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currency.NewBalanceStateValueMerger(height, tgBalSt.Key(), fact.Currency(), st)
				},
			))
		}
	}

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
