package contract

import (
	"context"
	"sync"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	cestate "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	currencytypes "github.com/ProtoconNet/mitum-currency/v3/types"
	ptypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

var registerContractProcessorPool = sync.Pool{
	New: func() interface{} {
		return new(RegisterContractProcessor)
	},
}

var initializeFuncName = "Initialize"

type RegisterContractProcessor struct {
	*base.BaseOperationProcessor
	encs *encoder.Encoders
}

func NewRegisterContractProcessor(encs encoder.Encoders) currencytypes.GetNewProcessor {
	return func(
		height base.Height,
		getStateFunc base.GetStateFunc,
		newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
		newProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	) (base.OperationProcessor, error) {
		e := util.StringError("failed to create new RegisterContractProcessor")

		nopp := registerContractProcessorPool.Get()
		opp, ok := nopp.(*RegisterContractProcessor)
		if !ok {
			return nil, errors.Errorf("expected RegisterContractProcessor, not %T", nopp)
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

func (opp *RegisterContractProcessor) PreProcess(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc,
) (context.Context, base.OperationProcessReasonError, error) {
	fact, ok := op.Fact().(RegisterContractFact)
	if !ok {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Wrap(common.ErrMTypeMismatch).
				Errorf("expected %T, not %T", RegisterContractFact{}, op.Fact())), nil
	}

	if err := fact.IsValid(nil); err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	_, cSt, aErr, cErr := cstate.ExistsCAccount(fact.Contract(), "contract", true, true, getStateFunc)
	if aErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Errorf("%v", aErr)), nil
	} else if cErr != nil {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Errorf("%v", cErr)), nil
	}

	ca, err := cestate.CheckCAAuthFromState(cSt, fact.Sender())
	if err != nil {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Errorf("%v", err)), nil
	}

	if ca == nil {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Wrap(common.ErrMValueInvalid).Errorf(
				"contract account value is nil")), nil
	}

	if ca.IsActive() {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceE).Errorf(
				"contract account %v has already been activated", fact.Contract())), nil
	}

	if found, _ := cstate.CheckNotExistsState(pstate.DesignStateKey(fact.Contract()), getStateFunc); found {
		return ctx, base.NewBaseOperationProcessReasonError("%s",
			common.ErrMPreProcess.
				Wrap(common.ErrMServiceE).Errorf("wasm service for contract account %v",
				fact.Contract(),
			)), nil
	}

	return ctx, nil, nil
}

func (opp *RegisterContractProcessor) Process(
	ctx context.Context, op base.Operation, getStateFunc base.GetStateFunc) (
	[]base.StateMergeValue, base.OperationProcessReasonError, error,
) {
	fact, _ := op.Fact().(RegisterContractFact)
	var sts []base.StateMergeValue

	schema, err := contractEngine.ValidateContract(fact.ContractCode())
	if err != nil {
		return nil, err, nil
	}

	execResult, bErr := contractEngine.ExecuteContract(
		*opp.encs,
		getStateFunc,
		cruntime.ExecuteRequest{
			Mode:         cruntime.InvocationModeRegister,
			Contract:     fact.Contract(),
			Sender:       fact.Sender(),
			Height:       opp.Height(),
			ContractCode: fact.ContractCode(),
			Schema:       &schema,
			Function:     initializeFuncName,
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

	design := ptypes.NewDesign(fact.ContractCode())
	if err := design.IsValid(nil); err != nil {
		return nil, base.NewBaseOperationProcessReasonError("invalid contract design, %q; %w", fact.Contract(), err), nil
	}

	persistedSchema := cruntime.NewPersistedContractSchema(fact.ContractCode(), schema)
	sts = append(sts, cstate.NewStateMergeValue(
		pstate.DesignStateKey(fact.Contract()),
		pstate.NewDesignStateValueWithSchema(design, &persistedSchema),
	))

	st, _ := cstate.ExistsState(cestate.StateKeyContractAccount(fact.Contract()), "contract account", getStateFunc)
	ca, _ := cestate.StateContractAccountValue(st)
	nca := ca.SetIsActive(true)

	sts = append(sts, cstate.NewStateMergeValue(
		cestate.StateKeyContractAccount(fact.Contract()),
		cestate.NewContractAccountStateValue(nca),
	))

	return sts, nil, nil
}

func (opp *RegisterContractProcessor) Close() error {
	registerContractProcessorPool.Put(opp)

	return nil
}
