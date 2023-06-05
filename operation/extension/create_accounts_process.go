package extension

//
//import (
//	"context"
//	"github.com/ProtoconNet/mitum-currency/v3/base"
//	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
//	"github.com/ProtoconNet/mitum-currency/v3/state"
//	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
//	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
//	"sync"
//
//	mitumbase "github.com/ProtoconNet/mitum2/base"
//	"github.com/ProtoconNet/mitum2/isaac"
//	"github.com/ProtoconNet/mitum2/util"
//	"github.com/pkg/errors"
//)
//
//var createAccountsItemProcessorPool = sync.Pool{
//	New: func() interface{} {
//		return new(CreateAccountsItemProcessor)
//	},
//}
//
//var createAccountsProcessorPool = sync.Pool{
//	New: func() interface{} {
//		return new(CreateAccountsProcessor)
//	},
//}
//
//type CreateAccountsItemProcessor struct {
//	h    util.Hash
//	item currency.CreateAccountsItem
//	ns   mitumbase.StateMergeValue
//	nb   map[base.CurrencyID]mitumbase.StateMergeValue
//}
//
//func (opp *CreateAccountsItemProcessor) PreProcess(
//	_ context.Context, _ mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
//) error {
//	for i := range opp.item.Amounts() {
//		am := opp.item.Amounts()[i]
//
//		policy, err := state.ExistsCurrencyPolicy(am.Currency(), getStateFunc)
//		if err != nil {
//			return err
//		}
//
//		if am.Big().Compare(policy.NewAccountMinBalance()) < 0 {
//			return errors.Errorf("amount should be over minimum balance, %v < %v", am.Big(), policy.NewAccountMinBalance())
//		}
//	}
//
//	target, err := opp.item.Address()
//	if err != nil {
//		return err
//	}
//
//	st, err := state.NotExistsState(statecurrency.StateKeyAccount(target), "key of target account", getStateFunc)
//	if err != nil {
//		return err
//	}
//	opp.ns = state.NewStateMergeValue(st.Key(), st.Value())
//
//	nb := map[base.CurrencyID]mitumbase.StateMergeValue{}
//	for i := range opp.item.Amounts() {
//		am := opp.item.Amounts()[i]
//		switch _, found, err := getStateFunc(statecurrency.StateKeyBalance(target, am.Currency())); {
//		case err != nil:
//			return err
//		case found:
//			return isaac.ErrStopProcessingRetry.Errorf("target balance already exists, %q", target)
//		default:
//			nb[am.Currency()] = state.NewStateMergeValue(statecurrency.StateKeyBalance(target, am.Currency()), statecurrency.NewBalanceStateValue(base.NewZeroAmount(am.Currency())))
//		}
//	}
//	opp.nb = nb
//
//	return nil
//}
//
//func (opp *CreateAccountsItemProcessor) Process(
//	_ context.Context, _ mitumbase.Operation, _ mitumbase.GetStateFunc,
//) ([]mitumbase.StateMergeValue, error) {
//	e := util.StringErrorFunc("failed to preprocess for CreateAccountsItemProcessor")
//
//	var (
//		nac base.Account
//		err error
//	)
//
//	if opp.item.AddressType() == base.EthAddressHint.Type() {
//		nac, err = base.NewEthAccountFromKeys(opp.item.Keys())
//	} else {
//		nac, err = base.NewAccountFromKeys(opp.item.Keys())
//	}
//	if err != nil {
//		return nil, e(err, "")
//	}
//
//	sts := make([]mitumbase.StateMergeValue, len(opp.item.Amounts())+1)
//	sts[0] = state.NewStateMergeValue(opp.ns.Key(), statecurrency.NewAccountStateValue(nac))
//
//	for i := range opp.item.Amounts() {
//		am := opp.item.Amounts()[i]
//		v, ok := opp.nb[am.Currency()].Value().(statecurrency.BalanceStateValue)
//		if !ok {
//			return nil, errors.Errorf("expected BalanceStateValue, not %T", opp.nb[am.Currency()].Value())
//		}
//		stv := statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Add(am.Big())))
//		sts[i+1] = state.NewStateMergeValue(opp.nb[am.Currency()].Key(), stv)
//	}
//
//	return sts, nil
//}
//
//func (opp *CreateAccountsItemProcessor) Close() error {
//	opp.h = nil
//	opp.item = nil
//	opp.ns = nil
//	opp.nb = nil
//
//	createAccountsItemProcessorPool.Put(opp)
//
//	return nil
//}
//
//type CreateAccountsProcessor struct {
//	*mitumbase.BaseOperationProcessor
//}
//
//func NewCreateAccountsProcessor() GetNewProcessor {
//	return func(
//		height mitumbase.Height,
//		getStateFunc mitumbase.GetStateFunc,
//		newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
//		newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
//	) (mitumbase.OperationProcessor, error) {
//		e := util.StringErrorFunc("failed to create new CreateAccountsProcessor")
//
//		nopp := createAccountsProcessorPool.Get()
//		opp, ok := nopp.(*CreateAccountsProcessor)
//		if !ok {
//			return nil, e(nil, "expected CreateAccountsProcessor, not %T", nopp)
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
//func (opp *CreateAccountsProcessor) PreProcess(
//	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc,
//) (context.Context, mitumbase.OperationProcessReasonError, error) {
//	e := util.StringErrorFunc("failed to preprocess CreateAccounts")
//
//	fact, ok := op.Fact().(currency.CreateAccountsFact)
//	if !ok {
//		return ctx, nil, e(nil, "expected CreateAccountsFact, not %T", op.Fact())
//	}
//
//	if err := state.CheckExistsState(statecurrency.StateKeyAccount(fact.Sender()), getStateFunc); err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("sender not found, %q: %w", fact.Sender(), err), nil
//	}
//
//	if err := state.CheckNotExistsState(extension.StateKeyContractAccount(fact.Sender()), getStateFunc); err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("contract account cannot be create-account sender, %q: %w", fact.Sender(), err), nil
//	}
//
//	if err := state.CheckFactSignsByState(fact.Sender(), op.Signs(), getStateFunc); err != nil {
//		return ctx, mitumbase.NewBaseOperationProcessReasonError("invalid signing: %w", err), nil
//	}
//
//	return ctx, nil, nil
//}
//
//func (opp *CreateAccountsProcessor) Process( // nolint:dupl
//	ctx context.Context, op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (
//	[]mitumbase.StateMergeValue, mitumbase.OperationProcessReasonError, error,
//) {
//	e := util.StringErrorFunc("failed to process CreateAccounts")
//
//	fact, ok := op.Fact().(currency.CreateAccountsFact)
//	if !ok {
//		return nil, nil, e(nil, "expected CreateAccountsFact, not %T", op.Fact())
//	}
//
//	required, err := opp.calculateItemsFee(op, getStateFunc)
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to calculate fee: %w", err), nil
//	}
//
//	sb, err := CheckEnoughBalance(fact.Sender(), required, getStateFunc)
//	if err != nil {
//		return nil, mitumbase.NewBaseOperationProcessReasonError("failed to check enough balance: %w", err), nil
//	}
//
//	ns := make([]*CreateAccountsItemProcessor, len(fact.Items()))
//	for i := range fact.Items() {
//		cip := createAccountsItemProcessorPool.Get()
//		c, ok := cip.(*CreateAccountsItemProcessor)
//		if !ok {
//			return nil, nil, e(nil, "expected CreateAccountsItemProcessor, not %T", cip)
//		}
//
//		c.h = op.Hash()
//		c.item = fact.Items()[i]
//
//		if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
//			return nil, mitumbase.NewBaseOperationProcessReasonError("fail to preprocess CreateAccountsItem: %w", err), nil
//		}
//
//		ns[i] = c
//	}
//
//	var sts []mitumbase.StateMergeValue // nolint:prealloc
//	for i := range ns {
//		s, err := ns[i].Process(ctx, op, getStateFunc)
//		if err != nil {
//			return nil, mitumbase.NewBaseOperationProcessReasonError("failed to process CreateAccountsItem: %w", err), nil
//		}
//		sts = append(sts, s...)
//
//		ns[i].Close()
//	}
//
//	for i := range sb {
//		v, ok := sb[i].Value().(statecurrency.BalanceStateValue)
//		if !ok {
//			return nil, nil, e(nil, "expected BalanceStateValue, not %T", sb[i].Value())
//		}
//		stv := statecurrency.NewBalanceStateValue(v.Amount.WithBig(v.Amount.Big().Sub(required[i][0])))
//		sts = append(sts, state.NewStateMergeValue(sb[i].Key(), stv))
//	}
//
//	return sts, nil, nil
//}
//
//func (opp *CreateAccountsProcessor) Close() error {
//	createAccountsProcessorPool.Put(opp)
//
//	return nil
//}
//
//func (opp *CreateAccountsProcessor) calculateItemsFee(op mitumbase.Operation, getStateFunc mitumbase.GetStateFunc) (map[base.CurrencyID][2]base.Big, error) {
//	fact, ok := op.Fact().(currency.CreateAccountsFact)
//	if !ok {
//		return nil, errors.Errorf("expected CreateAccountsFact, not %T", op.Fact())
//	}
//
//	items := make([]currency.AmountsItem, len(fact.Items()))
//	for i := range fact.Items() {
//		items[i] = fact.Items()[i]
//	}
//
//	return CalculateItemsFee(getStateFunc, items)
//}
//
//func CalculateItemsFee(getStateFunc mitumbase.GetStateFunc, items []currency.AmountsItem) (map[base.CurrencyID][2]base.Big, error) {
//	required := map[base.CurrencyID][2]base.Big{}
//
//	for i := range items {
//		it := items[i]
//
//		for j := range it.Amounts() {
//			am := it.Amounts()[j]
//
//			rq := [2]base.Big{base.ZeroBig, base.ZeroBig}
//			if k, found := required[am.Currency()]; found {
//				rq = k
//			}
//
//			policy, err := state.ExistsCurrencyPolicy(am.Currency(), getStateFunc)
//			if err != nil {
//				return nil, err
//			}
//
//			switch k, err := policy.Feeer().Fee(am.Big()); {
//			case err != nil:
//				return nil, err
//			case !k.OverZero():
//				required[am.Currency()] = [2]base.Big{rq[0].Add(am.Big()), rq[1]}
//			default:
//				required[am.Currency()] = [2]base.Big{rq[0].Add(am.Big()).Add(k), rq[1].Add(k)}
//			}
//		}
//	}
//
//	return required, nil
//}
//
//func CheckEnoughBalance(
//	holder mitumbase.Address,
//	required map[base.CurrencyID][2]base.Big,
//	getStateFunc mitumbase.GetStateFunc,
//) (map[base.CurrencyID]mitumbase.StateMergeValue, error) {
//	sb := map[base.CurrencyID]mitumbase.StateMergeValue{}
//
//	for cid := range required {
//		rq := required[cid]
//
//		st, err := state.ExistsState(statecurrency.StateKeyBalance(holder, cid), "key of holder balance", getStateFunc)
//		if err != nil {
//			return nil, err
//		}
//
//		am, err := statecurrency.StateBalanceValue(st)
//		if err != nil {
//			return nil, errors.Errorf("not enough balance of sender, %q: %w", holder, err)
//		}
//
//		if am.Big().Compare(rq[0]) < 0 {
//			return nil, errors.Errorf("not enough balance of sender, %q; %v !> %v", holder, am.Big(), rq[0])
//		}
//		sb[cid] = state.NewStateMergeValue(st.Key(), statecurrency.NewBalanceStateValue(am))
//	}
//
//	return sb, nil
//}
