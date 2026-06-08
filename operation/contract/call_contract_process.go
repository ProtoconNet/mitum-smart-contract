package contract

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract/runtime"
	"github.com/ProtoconNet/mitum-smart-contract/state"
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
	proposal *base.ProposalSignFact
	encs     *encoder.Encoders
}

func NewCallContractProcessor(encs encoder.Encoders) ctypes.GetNewProcessorWithProposal {
	return func(
		height base.Height,
		proposal *base.ProposalSignFact,
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
		opp.proposal = proposal

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
	designState, found, err := getStateFunc(state.DesignStateKey(fact.Contract()))
	switch {
	case err != nil:
		return base.NewBaseOperationProcessReasonError("failed to read design state: %v", err)
	case !found:
		return base.NewBaseOperationProcessReasonError("contract design not found for typed contract %v", fact.Contract())
	}
	if _, err := state.GetDesignStateValueFromState(designState); err != nil {
		return base.NewBaseOperationProcessReasonError("failed to decode design state: %v", err)
	}

	runtimeState, found, err := getStateFunc(state.RuntimeStateKey(fact.Contract()))
	switch {
	case err != nil:
		return base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	case !found:
		return base.NewBaseOperationProcessReasonError("runtime state not found for typed contract %v", fact.Contract())
	}
	runtimeValue, err := state.GetRuntimeFromState(runtimeState)
	if err != nil {
		return base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
	}
	if runtimeValue.Engine != state.RuntimeEngineGnoSnapshot {
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

	blockTime, err := proposalBlockTime(opp.proposal)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"failed to resolve contract write block time: %v", err,
		), nil
	}

	if len(fact.items) < 1 {
		return nil, base.NewBaseOperationProcessReasonError(
			"missing call items"), nil
	}

	st, err := cstate.ExistsState(state.DesignStateKey(fact.Contract()), "contract design", getStateFunc)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"%v", err), nil
	}
	dsv, err := state.GetDesignStateValueFromState(st)
	if err != nil {
		return nil, base.NewBaseOperationProcessReasonError(
			"%v", err), nil
	}
	cd := dsv.Design

	var schema *runtime.ContractSchema
	if persistedSchema, ok := runtime.RuntimeSchemaFromPersisted(cd.ContractCode(), dsv.Schema); ok {
		schema = &persistedSchema
	}

	execResult, bErr := contractEngine.ExecuteContract(
		*opp.encs,
		getStateFunc,
		runtime.ExecuteRequest{
			Mode:         runtime.InvocationModeCall,
			Contract:     fact.Contract(),
			Sender:       fact.Sender(),
			Height:       opp.Height(),
			BlockTime:    blockTime,
			ContractCode: cd.ContractCode(),
			Schema:       schema,
			CallItems:    fact.runtimeCallItems(),
		},
	)
	if bErr != nil {
		return nil, bErr, nil
	}

	if execResult.Engine != state.RuntimeEngineGnoSnapshot {
		return nil, base.NewBaseOperationProcessReasonError(
			"unsupported runtime engine %q", execResult.Engine,
		), nil
	}
	sts = append(sts, execResult.StateMerges...)

	return sts, nil, nil
}

func (opp *CallContractProcessor) Close() error {
	opp.proposal = nil
	callContractProcessorPool.Put(opp)

	return nil
}
