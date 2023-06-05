package extension

//
//import (
//	"context"
//	"github.com/ProtoconNet/mitum-currency/v3/base"
//	"github.com/ProtoconNet/mitum-currency/v3/state"
//	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
//	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
//	"sync"
//
//	mitumbase "github.com/ProtoconNet/mitum2/base"
//	"github.com/ProtoconNet/mitum2/util"
//	"github.com/pkg/errors"
//)
//
//var keyUpdaterProcessorPool = sync.Pool{
//	New: func() interface{} {
//		return new(KeyUpdaterProcessor)
//	},
//}
//
//type KeyUpdaterProcessor struct {
//	*mitumbase.BaseOperationProcessor
//}
//
//func NewKeyUpdaterProcessor() GetNewProcessor {
//	return func(
//		height mitumbase.Height,
//		getStateFunc mitumbase.GetStateFunc,
//		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
//		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
//	) (mitumbase.OperationProcessor, error) {
//		e := util.StringErrorFunc("failed to create new KeyUpdaterProcessor")
//
//		nopp := keyUpdaterProcessorPool.Get()
//		opp, ok := nopp.(*KeyUpdaterProcessor)
//		if !ok {
//			return nil, errors.Errorf("expected KeyUpdaterProcessor, not %T", nopp)
//		}
//
//		b, err := mitumbase.NewBaseOperationProcessor(
//			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
//		if err != nil {
//			return nil, e(err, "")
//		}
//
//		opp.BaseOperationProcessor = b
//
//		return opp, nil
//	}
//}
//
//func (opp *KeyUpdaterProcessor) PreProcess(
//	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
//) (context.Context, mitumbase.OperationProcessReasonError, error) {
//	e := util.StringErrorFunc("failed to preprocess KeyUpdater")
//
//	fact, ok := op.Fact().(KeyUpdaterFact)
//	if !ok {
//		return ctx, nil, e(nil, "expected KeyUpdaterFact, not %T", op.Fact())
//	}
//
//	st, err := state.ExistsState(currency.StateKeyAccount(fact.Target()), "key of target account", getStateFunc)
//	if err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("target not found, %q: %w", fact.Target(), err), nil
//	}
//
//	if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.Target()), getStateFunc); err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account not allowed for key updater, %q: %w", fact.Target(), err), nil
//	}
//
//	ks, err := currency.StateKeysValue(st)
//	if err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("failed to get keys value, %q: %w", fact.Keys().Hash(), err), nil
//	}
//	if ks.Equal(fact.Keys()) {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("same Keys as existing, %q: %w", fact.Keys().Hash(), err), nil
//	}
//
//	if err := state.CheckFactSignsByState(fact.Target(), op.Signs(), getStateFunc); err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
//	}
//
//	return ctx, nil, nil
//}
//
//func (opp *KeyUpdaterProcessor) Process( // nolint:dupl
//	_ context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
//	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
//) {
//	e := util.StringErrorFunc("failed to process KeyUpdater")
//
//	fact, ok := op.Fact().(KeyUpdaterFact)
//	if !ok {
//		return nil, nil, e(nil, "expected KeyUpdaterFact, not %T", op.Fact())
//	}
//
//	st, err := state.ExistsState(currency.StateKeyAccount(fact.Target()), "key of target account", getStateFunc)
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("target not found, %q: %w", fact.Target(), err), nil
//	}
//	sa := state.NewStateMergeValue(st.Key(), st.Value())
//
//	policy, err := state.ExistsCurrencyPolicy(fact.Currency(), getStateFunc)
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("currency not found, %q: %w", fact.Currency(), err), nil
//	}
//	fee, err := policy.Feeer().Fee(base.ZeroBig)
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check fee of currency, %q: %w", fact.Currency(), err), nil
//	}
//
//	st, err = state.ExistsState(currency.StateKeyBalance(fact.Target(), fact.Currency()), "key of target balance", getStateFunc)
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("target balance not found, %q: %w", fact.Target(), err), nil
//	}
//	sb := state.NewStateMergeValue(st.Key(), st.Value())
//
//	switch b, err := currency.StateBalanceValue(st); {
//	case err != nil:
//		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to get balance value, %q: %w", currency.StateKeyBalance(fact.Target(), fact.Currency()), err), nil
//	case b.Big().Compare(fee) < 0:
//		return nil, mitumbase.NewBaseOperationProcessReasonError("not enough balance of target, %q", fact.Target()), nil
//	}
//
//	var sts []mitumbase.StateMergeValue // nolint:prealloc
//
//	v, ok := sb.Value().(currency.BalanceStateValue)
//	if !ok {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("expected BalanceStateValue, not %T", sb.Value()), nil
//	}
//	sts = append(sts, state.NewStateMergeValue(sb.Key(), currency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(fee)))))
//
//	a, err := base.NewAccountFromKeys(fact.Keys())
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to create new account from keys"), nil
//	}
//	sts = append(sts, state.NewStateMergeValue(sa.Key(), currency.NewAccountStateValue(a)))
//
//	return sts, nil, nil
//}
//
//func (opp *KeyUpdaterProcessor) Close() error {
//	keyUpdaterProcessorPool.Put(opp)
//
//	return nil
//}
