package contract

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
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
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", CallContractFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if reason := opp.preProcessStateGate(fact, getStateFunc); reason != nil {
		return ctx, reason, nil
	}

	return ctx, nil, nil
}

func (opp *CallContractProcessor) preProcessStateGate(
	fact CallContractFact,
	getStateFunc base.GetStateFunc,
) base.OperationProcessReasonError {
	designState, found, err := getStateFunc(pstate.DesignStateKey(fact.Contract()))
	switch {
	case err != nil:
		return base.NewBaseOperationProcessReasonError("failed to read design state: %v", err)
	case !found:
		return base.NewBaseOperationProcessReasonError("contract design not found for typed contract %v", fact.Contract())
	}
	if _, err := pstate.GetDesignStateValueFromState(designState); err != nil {
		return base.NewBaseOperationProcessReasonError("failed to decode design state: %v", err)
	}

	runtimeState, found, err := getStateFunc(pstate.RuntimeStateKey(fact.Contract()))
	switch {
	case err != nil:
		return base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	case !found:
		return base.NewBaseOperationProcessReasonError("runtime state not found for typed contract %v", fact.Contract())
	}
	runtimeValue, err := pstate.GetRuntimeFromState(runtimeState)
	if err != nil {
		return base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
	}
	if runtimeValue.Engine != pstate.RuntimeEngineGnoSnapshot {
		return base.NewBaseOperationProcessReasonError("unsupported runtime engine %q", runtimeValue.Engine)
	}

	return nil
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
	dsv, err := pstate.GetDesignStateValueFromState(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"%v", err), nil
	}
	cd := dsv.Design

	var schema *cruntime.ContractSchema
	if persistedSchema, ok := cruntime.RuntimeSchemaFromPersisted(cd.ContractCode(), dsv.Schema); ok {
		schema = &persistedSchema
	}

	execResult, bErr := contractEngine.ExecuteContract(
		*opp.encs,
		getStateFunc,
		cruntime.ExecuteRequest{
			Mode:         cruntime.InvocationModeCall,
			Contract:     fact.Contract(),
			Sender:       fact.Sender(),
			Height:       opp.Height(),
			ContractCode: cd.ContractCode(),
			Schema:       schema,
			Function:     fName,
			CallData:     fact.callData,
		},
	)
	if bErr != nil {
		return nil, bErr, nil
	}

	if execResult.Engine != pstate.RuntimeEngineGnoSnapshot {
		return nil, base.NewBaseOperationProcessReasonError(
			"unsupported runtime engine %q", execResult.Engine,
		), nil
	}
	sts = append(sts, execResult.StateMerges...)

	return sts, nil, nil
}

func (opp *CallContractProcessor) Close() error {
	callContractProcessorPool.Put(opp)

	return nil
}
