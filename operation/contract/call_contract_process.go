package contract

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

var callContractProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(CallContractProcessor)
	},
}

type CallContractProcessor struct {
	*base.BaseOperationProcessor
	encs *encoder.Encoders
}

func NewCallContractProcessor(encs encoder.Encoders) currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new CallContractProcessor")

		nopp := callContractProcessorPool.Get()
		opp, ok := nopp.(*CallContractProcessor)
		if !ok {
			return nil, errors.Errorf("expected CallContractProcessor, not %T", nopp)
		}

		b, err := base.NewBaseOperationProcessor(
			height, getStateFunc, newPreProcessConstraintFunc, newProcessConstraintFunc)
		if err != nil {
			return nil, e.Wrap(err)
		}

		if opp.encs == nil {
			opp.encs = &encs
		}

		opp.BaseOperationProcessor = b

		return opp, nil
	}
}

func (opp *CallContractProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(CallContractFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", CallContractFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError(
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	return ctx, nil, nil
}

func (opp *CallContractProcessor) Process(
	_ context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(CallContractFact)

	var sts []base.StateMergeValue

	fName, found := fact.callData["function"]
	if !found {
		return nil, base.NewBaseOperationProcessReasonError(
			"missing function name in call data"), nil
	}

	st, err := cstate.ExistsState(pstate.DesignStateKey(fact.Contract()), "contract design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"%v", err), nil
	}
	cd, err := pstate.GetDesignFromState(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"%v", err), nil
	}

	results, bErr := ExecuteContract(
		*opp.encs, getStateFunc, fact.Contract(), fact.Sender(), cd.ContractCode(), fName, fact.callData,
	)
	if bErr != nil {
		return nil, bErr, nil
	}

	var result map[string]interface{}
	var ok bool
	if !results[0].IsNil() {
		result, ok = results[0].Interface().(map[string]interface{})
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"%v function must return map[string]interface{}, but got %T; %v", fName, results[0].Interface()), nil
		}
	}

	if !results[1].IsNil() {
		err, ok = results[1].Interface().(error)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"%v function did not return an error as expected, got %T", fName, results[1].Interface()), nil
		}
	}

	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to initialize contract code at %v; %v", fact.Contract(), err), nil
	}
	if result != nil {
		key, found := result["key"]
		if !found {
			return nil, base.NewBaseOperationProcessReasonError(
				"key not found from Initialize result of contract code at %v", fact.Contract()), nil
		}
		stKey, ok := key.(string)
		if !ok {
			return nil, base.NewBaseOperationProcessReasonError(
				"key type expected string, but %T", key), nil
		}
		sts = append(sts, cstate.NewStateMergeValue(
			pstate.DataStateKey(fact.Contract(), stKey),
			pstate.NewDataStateValue(result),
		))
	}

	return sts, nil, nil
}

func (opp *CallContractProcessor) Close() error {
	callContractProcessorPool.Put(opp)

	return nil
}
