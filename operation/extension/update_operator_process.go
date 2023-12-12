package extension

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"sync"

	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var UpdateOperatorProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(UpdateOperatorProcessor)
	},
}

func (UpdateOperator) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type UpdateOperatorProcessor struct {
	*base.BaseOperationProcessor
	ca  base.StateMergeValue
	sb  base.StateMergeValue
	fee common.Big
}

func NewUpdateOperatorProcessor() types.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("create new UpdateOperatorProcessor")

		nopp := UpdateOperatorProcessorPool.Get()
		opp, ok := nopp.(*UpdateOperatorProcessor)
		if !ok {
			return nil, errors.Errorf("expected UpdateOperatorProcessor, not %T", nopp)
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

func (opp *UpdateOperatorProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(UpdateOperatorFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError("expected UpdateOperatorFact, not %T", op.Fact()), nil
	}

	if err := state.CheckFactSignsByState(fact.sender, op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("invalid signing; %w", err), nil
	}

	if err := state.CheckExistsState(currency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("check existence of sender %v; %w", fact.Sender(), err), nil
	} else if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("contract account cannot update operator, %v; %w", fact.Sender(), err), nil
	} else if err = state.CheckExistsState(currency.StateKeyAccount(fact.Contract()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("check existence of contract %v; %w", fact.Contract(), err), nil
	} else if err := state.CheckExistsState(extension.StateKeyContractAccount(fact.Contract()), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("check existence of target contract account %v; %w", fact.Contract(), err), nil
	}

	for i := range fact.Operators() {
		if err := state.CheckExistsState(currency.StateKeyAccount(fact.Operators()[i]), getStateFunc); err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("check existence of sender %v; %w", fact.Operators()[i], err), nil
		} else if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.Operators()[i]), getStateFunc); err != nil {
			return ctx, base.NewBaseOperationProcessReasonError("contract account cannot be an operator, %v; %w", fact.Sender(), err), nil
		}
	}

	return ctx, nil, nil
}

func (opp *UpdateOperatorProcessor) Process( // nolint:dupl
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	e := util.StringError("process UpdateOperator")

	fact, ok := op.Fact().(UpdateOperatorFact)
	if !ok {
		return nil, nil, e.Errorf("expected UpdateOperatorFact, not %T", op.Fact())
	}

	var ctAccSt base.State
	var err error
	ctAccSt, err = state.ExistsState(extension.StateKeyContractAccount(fact.Contract()), "contract account status", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check existence of contract account status %v ; %w", fact.Contract(), err), nil
	}

	var fee common.Big
	var policy types.CurrencyPolicy
	if policy, err = state.ExistsCurrencyPolicy(fact.Currency(), getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check existence of currency %v; %w", fact.Currency(), err), nil
	} else if fee, err = policy.Feeer().Fee(common.ZeroBig); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check fee of currency %v; %w", fact.Currency(), err), nil
	}

	var sdBalSt base.State
	if sdBalSt, err = state.ExistsState(currency.StateKeyBalance(fact.Sender(), fact.currency), "balance of sender", getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check existence of sender balance %v ; %w", fact.Sender(), err), nil
	} else if b, err := currency.StateBalanceValue(sdBalSt); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check existence of sender balance %v, %v ; %w", fact.currency, fact.Sender(), err), nil
	} else if b.Big().Compare(fee) < 0 {
		return nil, base.NewBaseOperationProcessReasonError("insufficient balance with fee %v ,%v", fact.currency, fact.Sender()), nil
	}

	var stmvs []base.StateMergeValue // nolint:prealloc
	v, ok := sdBalSt.Value().(currency.BalanceStateValue)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", sdBalSt.Value()), nil
	}

	if policy.Feeer().Receiver() != nil {
		if err := state.CheckExistsState(currency.StateKeyAccount(policy.Feeer().Receiver()), getStateFunc); err != nil {
			return nil, nil, err
		} else if feeRcvrSt, found, err := getStateFunc(currency.StateKeyBalance(policy.Feeer().Receiver(), fact.currency)); err != nil {
			return nil, nil, err
		} else if !found {
			return nil, nil, errors.Errorf("feeer receiver %s not found", policy.Feeer().Receiver())
		} else if feeRcvrSt.Key() != sdBalSt.Key() {
			r, ok := feeRcvrSt.Value().(currency.BalanceStateValue)
			if !ok {
				return nil, nil, errors.Errorf("invalid BalanceState value found, %T", feeRcvrSt.Value())
			}
			stmvs = append(stmvs, common.NewBaseStateMergeValue(
				feeRcvrSt.Key(),
				currency.NewAddBalanceStateValue(r.Amount.WithBig(fee)),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currency.NewBalanceStateValueMerger(height, feeRcvrSt.Key(), fact.currency, st)
				},
			))

			stmvs = append(stmvs, common.NewBaseStateMergeValue(
				sdBalSt.Key(),
				currency.NewDeductBalanceStateValue(v.Amount.WithBig(fee)),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currency.NewBalanceStateValueMerger(height, sdBalSt.Key(), fact.currency, st)
				},
			))
		}
	}

	ctsv := ctAccSt.Value()
	if ctsv == nil {
		return nil, nil, util.ErrNotFound.Errorf("contract account status not found in State")
	}

	sv, ok := ctsv.(extension.ContractAccountStateValue)
	if !ok {
		return nil, nil, errors.Errorf("invalid contract account value found, %T", ctsv)
	}

	status := sv.Status()
	err = status.SetOperators(fact.Operators())
	if err != nil {
		return nil, nil, err
	}

	stmvs = append(stmvs, state.NewStateMergeValue(ctAccSt.Key(), extension.NewContractAccountStateValue(status)))

	return stmvs, nil, nil
}

func (opp *UpdateOperatorProcessor) Close() error {
	UpdateOperatorProcessorPool.Put(opp)

	return nil
}
