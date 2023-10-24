package currency

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"

	"github.com/ProtoconNet/mitum2/isaac"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

var registerCurrencyProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RegisterCurrencyProcessor)
	},
}

func (RegisterCurrency) Process(
	_ context.Context, _ base.GetStateFunc,
) ([]base.StateMergeValue, base.OperationProcessReasonError, error) {
	// NOTE Process is nil func
	return nil, nil, nil
}

type RegisterCurrencyProcessor struct {
	*base.BaseOperationProcessor
	suffrage  base.Suffrage
	threshold base.Threshold
}

func NewRegisterCurrencyProcessor(threshold base.Threshold) types.GetNewProcessor {
	return func(height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("create new RegisterCurrencyProcessor")

		nopp := registerCurrencyProcessorPool.Get()
		opp, ok := nopp.(*RegisterCurrencyProcessor)
		if !ok {
			return nil, e.Wrap(errors.Errorf("expected %T, not %T", &RegisterCurrencyProcessor{}, nopp))
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		opp.BaseOperationProcessor = b
		opp.threshold = threshold

		switch i, found, err := getStateFunc(isaac.SuffrageStateKey); {
		case err != nil:
			return nil, e.Wrap(err)
		case !found, i == nil:
			return nil, e.Wrap(isaac.ErrStopProcessingRetry.Errorf("empty state"))
		default:
			sufstv := i.Value().(base.SuffrageNodesStateValue) //nolint:forcetypeassert //...

			suf, err := sufstv.Suffrage()
			if err != nil {
				return nil, e.Wrap(isaac.ErrStopProcessingRetry.Errorf("get suffrage from state"))
			}

			opp.suffrage = suf
		}

		return opp, nil
	}
}

func (opp *RegisterCurrencyProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	e := util.StringError("preprocess for RegisterCurrency")

	nop, ok := op.(RegisterCurrency)
	if !ok {
		return ctx, nil, e.Errorf("not RegisterCurrency, %T", op)
	}

	if err := base.CheckFactSignsBySuffrage(opp.suffrage, opp.threshold, nop.NodeSigns()); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("not enough signs"), nil
	}

	fact, ok := op.Fact().(RegisterCurrencyFact)
	if !ok {
		return ctx, nil, e.Errorf("not RegisterCurrencyFact, %T", op.Fact())
	}

	design := fact.currency

	_, err := state.NotExistsState(
		currency.StateKeyCurrencyDesign(design.Currency()),
		design.Currency().String(),
		getStateFunc,
	)
	if err != nil {
		return ctx, nil, err
	}

	if err := state.CheckExistsState(currency.StateKeyAccount(design.GenesisAccount()), getStateFunc); err != nil {
		return ctx, nil, e.WithMessage(err, "genesis account not found")
	}

	if receiver := design.Policy().Feeer().Receiver(); receiver != nil {
		if err := state.CheckExistsState(currency.StateKeyAccount(receiver), getStateFunc); err != nil {
			return ctx, nil, e.WithMessage(err, "feeer receiver account not found")
		}
	}

	switch _, found, err := getStateFunc(currency.StateKeyCurrencyDesign(design.Currency())); {
	case err != nil:
		return ctx, nil, err
	case found:
		return ctx, nil, base.NewBaseOperationProcessReasonError("currency already registered, %v", design.Currency())
	default:
	}

	switch _, found, err := getStateFunc(currency.StateKeyBalance(design.GenesisAccount(), design.Currency())); {
	case err != nil:
		return ctx, nil, err
	case found:
		return ctx, nil, base.NewBaseOperationProcessReasonError("genesis account has already the currency, %v", design.Currency())
	default:
	}

	return ctx, nil, nil
}

func (opp *RegisterCurrencyProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, ok := op.Fact().(RegisterCurrencyFact)
	if !ok {
		return nil, nil, errors.Errorf("expected %T, not %T", RegisterCurrencyFact{}, op.Fact())
	}

	sts := make([]base.StateMergeValue, 4)

	design := fact.currency

	ba := currency.NewBalanceStateValue(design.Amount())
	sts[0] = state.NewStateMergeValue(
		currency.StateKeyBalance(design.GenesisAccount(), design.Currency()),
		ba,
	)

	de := currency.NewCurrencyDesignStateValue(design)
	sts[1] = state.NewStateMergeValue(currency.StateKeyCurrencyDesign(design.Currency()), de)

	{
		l, err := createZeroAccount(design.Currency(), getStateFunc)
		if err != nil {
			return nil, nil, err
		}
		sts[2], sts[3] = l[0], l[1]
	}

	return sts, nil, nil
}

func createZeroAccount(
	cid types.CurrencyID,
	getStateFunc base.GetStateFunc,
) ([]base.StateMergeValue, error) {
	sts := make([]base.StateMergeValue, 2)

	ac, err := types.ZeroAccount(cid)
	if err != nil {
		return nil, err
	}
	ast, err := state.NotExistsState(currency.StateKeyAccount(ac.Address()), "keys of zero account", getStateFunc)
	if err != nil {
		return nil, err
	}

	sts[0] = state.NewStateMergeValue(ast.Key(), currency.NewAccountStateValue(ac))

	bst, err := state.NotExistsState(currency.StateKeyBalance(ac.Address(), cid), "balance of zero account", getStateFunc)
	if err != nil {
		return nil, err
	}

	sts[1] = state.NewStateMergeValue(bst.Key(), currency.NewBalanceStateValue(types.NewZeroAmount(cid)))

	return sts, nil
}

func (opp *RegisterCurrencyProcessor) Close() error {
	opp.suffrage = nil
	opp.threshold = 0

	registerCurrencyProcessorPool.Put(opp)

	return nil
}
