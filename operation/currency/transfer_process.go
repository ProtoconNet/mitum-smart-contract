package currency

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
	"sync"
)

var transferItemProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransferItemProcessor)
	},
}

var transferProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(TransferProcessor)
	},
}

func (Transfer) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type TransferItemProcessor struct {
	h    util.Hash
	item TransferItem
	rb   map[types.CurrencyID]base.StateMergeValue
}

func (opp *TransferItemProcessor) PreProcess(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) error {
	rb := map[types.CurrencyID]base.StateMergeValue{}
	amounts := opp.item.Amounts()
	for i := range amounts {
		am := amounts[i]
		cid := am.Currency()
		receiver := opp.item.Receiver()
		_, err := state.ExistsCurrencyPolicy(cid, getStateFunc)
		if err != nil {
			return err
		}

		st, _, err := getStateFunc(currency.BalanceStateKey(receiver, cid))
		if err != nil {
			return err
		}

		var balance types.Amount
		if st == nil {
			balance = types.NewZeroAmount(cid)
		} else {
			balance, err = currency.StateBalanceValue(st)
			if err != nil {
				return err
			}
		}

		rb[am.Currency()] = common.NewBaseStateMergeValue(
			currency.BalanceStateKey(receiver, cid),
			currency.NewAddBalanceStateValue(balance),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currency.NewBalanceStateValueMerger(
					height,
					currency.BalanceStateKey(receiver, cid),
					cid,
					st,
				)
			},
		)
	}

	opp.rb = rb

	return nil
}

func (opp *TransferItemProcessor) Process(
	_ context.Context, _ base.Operation, getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	e := util.StringError("preprocess TransferItemProcessor")

	var sts []base.StateMergeValue
	receiver := opp.item.Receiver()
	k := currency.AccountStateKey(receiver)
	switch _, found, err := getStateFunc(k); {
	case err != nil:
		return nil, e.Wrap(err)
	case !found:
		nilKys, err := types.NewNilAccountKeysFromAddress(receiver)
		if err != nil {
			return nil, e.Wrap(err)
		}
		acc, err := types.NewAccount(receiver, nilKys)
		if err != nil {
			return nil, e.Wrap(err)
		}

		sts = append(sts, state.NewStateMergeValue(k, currency.NewAccountStateValue(acc)))
	default:
	}

	amounts := opp.item.Amounts()
	for i := range amounts {
		am := amounts[i]
		cid := am.Currency()
		stv := opp.rb[cid]
		v, ok := stv.Value().(currency.AddBalanceStateValue)
		if !ok {
			return nil, errors.Errorf("not AddBalanceStateValue, %T", stv.Value())
		}
		sts = append(sts, common.NewBaseStateMergeValue(
			stv.Key(),
			currency.NewAddBalanceStateValue(v.Amount.WithBig(am.Big())),
			func(height base.Height, st base.State) base.StateValueMerger {
				return currency.NewBalanceStateValueMerger(height, stv.Key(), cid, st)
			},
		))
	}

	return sts, nil
}

func (opp *TransferItemProcessor) Close() {
	opp.h = nil
	opp.item = nil
	opp.rb = nil

	transferItemProcessorPool.Put(opp)
}

type TransferProcessor struct {
	*base.BaseOperationProcessor
	//ns       []*TransferItemProcessor
	required map[types.CurrencyID][2]common.Big
}

func NewTransferProcessor() types.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("create new TransferProcessor")

		nopp := transferProcessorPool.Get()
		opp, ok := nopp.(*TransferProcessor)
		if !ok {
			return nil, e.Wrap(errors.Errorf("expected TransferProcessor, not %T", nopp))
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b
		//opp.ns = nil
		opp.required = nil

		return opp, nil
	}
}

func (opp *TransferProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(TransferFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMTypeMismatch).Errorf("expected %T, not %T", TransferFact{}, op.Fact()),
		), nil
	}

	if _, _, aErr, cErr := state.ExistsCAccount(fact.Sender(), "sender", true, false, getStateFunc); aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMCAccountNA).Errorf("%v", cErr)), nil
	}

	if err := state.CheckFactSignsByState(fact.Sender(), op.Signs(), getStateFunc); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.Wrap(common.ErrMSignInvalid).Errorf("%v", err)), nil
	}

	var wg sync.WaitGroup
	errChan := make(chan *base.BaseOperationProcessReasonError, len(fact.items))
	for i := range fact.items {
		wg.Add(1)
		go func(item TransferItem) {
			defer wg.Done()
			tip := transferItemProcessorPool.Get()
			t, ok := tip.(*TransferItemProcessor)
			if !ok {
				err := base.NewBaseOperationProcessReasonError(
					common.ErrMPreProcess.Wrap(
						common.ErrMTypeMismatch).Errorf("expected %T, not %T", &TransferItemProcessor{}, tip))
				errChan <- &err
				return
			}

			t.h = op.Hash()
			t.item = item

			if err := t.PreProcess(ctx, op, getStateFunc); err != nil {
				err := base.NewBaseOperationProcessReasonError(common.ErrMPreProcess.Errorf("%v", err))
				errChan <- &err
				return
			}
			t.Close()
		}(fact.items[i])
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return nil, *err, nil
		}
	}

	return ctx, nil, nil
}

func (opp *TransferProcessor) Process( // nolint:dupl
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(TransferFact)
	if !ok {
		return nil, base.NewBaseOperationProcessReasonError("expected %T, not %T", TransferFact{}, op.Fact()), nil
	}

	var (
		senderBalSts, feeReceiverBalSts map[types.CurrencyID]base.State
		required                        map[types.CurrencyID][2]common.Big
		err                             error
	)

	if feeReceiverBalSts, required, err = opp.calculateItemsFee(op, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("calculate fee: %w", err), nil
	} else if senderBalSts, err = CheckEnoughBalance(fact.Sender(), required, getStateFunc); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("check enough balance: %w", err), nil
	} else {
		opp.required = required
	}

	//ns := make([]*TransferItemProcessor, len(fact.items))
	var stmvs []base.StateMergeValue // nolint:prealloc
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan *base.BaseOperationProcessReasonError, len(fact.items))
	for i := range fact.items {
		wg.Add(1)
		go func(item TransferItem) {
			defer wg.Done()
			cip := transferItemProcessorPool.Get()
			c, ok := cip.(*TransferItemProcessor)
			if !ok {
				err := base.NewBaseOperationProcessReasonError("expected %T, not %T", &TransferItemProcessor{}, cip)
				errChan <- &err
				return
			}

			c.h = op.Hash()
			c.item = fact.items[i]

			if err := c.PreProcess(ctx, op, getStateFunc); err != nil {
				err := base.NewBaseOperationProcessReasonError("fail to preprocess transfer item: %w", err)
				errChan <- &err
				return
			}

			s, err := c.Process(ctx, op, getStateFunc)
			if err != nil {
				err := base.NewBaseOperationProcessReasonError("process transfer item: %w", err)
				errChan <- &err
				return
			}
			mu.Lock()
			stmvs = append(stmvs, s...)
			mu.Unlock()
			c.Close()
		}(fact.items[i])
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		if err != nil {
			return nil, *err, nil
		}
	}

	for cid := range senderBalSts {
		v, ok := senderBalSts[cid].Value().(currency.BalanceStateValue)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"expected %T, not %T", currency.BalanceStateValue{}, senderBalSts[cid].Value()), nil
		}

		_, feeReceiverFound := feeReceiverBalSts[cid]

		var stmv base.StateMergeValue
		if feeReceiverFound && (senderBalSts[cid].Key() == feeReceiverBalSts[cid].Key()) {
			stmv = common.NewBaseStateMergeValue(
				senderBalSts[cid].Key(),
				currency.NewDeductBalanceStateValue(v.Amount.WithBig(opp.required[cid][0].Sub(opp.required[cid][1]))),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currency.NewBalanceStateValueMerger(height, senderBalSts[cid].Key(), cid, st)
				},
			)
		} else {
			stmv = common.NewBaseStateMergeValue(
				senderBalSts[cid].Key(),
				currency.NewDeductBalanceStateValue(v.Amount.WithBig(opp.required[cid][0])),
				func(height base.Height, st base.State) base.StateValueMerger {
					return currency.NewBalanceStateValueMerger(height, senderBalSts[cid].Key(), cid, st)
				},
			)
			if feeReceiverFound {
				r, ok := feeReceiverBalSts[cid].Value().(currency.BalanceStateValue)
				if !ok {
					return nil, base.NewBaseOperationProcessReasonError("expected %T, not %T", currency.BalanceStateValue{}, feeReceiverBalSts[cid].Value()), nil
				}
				stmvs = append(
					stmvs,
					common.NewBaseStateMergeValue(
						feeReceiverBalSts[cid].Key(),
						currency.NewAddBalanceStateValue(r.Amount.WithBig(opp.required[cid][1])),
						func(height base.Height, st base.State) base.StateValueMerger {
							return currency.NewBalanceStateValueMerger(height, feeReceiverBalSts[cid].Key(), cid, st)
						},
					),
				)
			}
		}
		stmvs = append(stmvs, stmv)
	}

	return stmvs, nil, nil
}

func (opp *TransferProcessor) Close() error {
	opp.required = nil

	transferProcessorPool.Put(opp)

	return nil
}

func (opp *TransferProcessor) calculateItemsFee(op base.Operation, getStateFunc base.GetStateFunc) (map[types.CurrencyID]base.State, map[types.CurrencyID][2]common.Big, error) {
	fact, ok := op.Fact().(TransferFact)
	if !ok {
		return nil, nil, errors.Errorf("expected %T, not %T", TransferFact{}, op.Fact())
	}
	items := make([]AmountsItem, len(fact.items))
	for i := range fact.items {
		items[i] = fact.items[i]
	}

	return CalculateItemsFee(getStateFunc, items)
}
