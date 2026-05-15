package runtime

import (
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type hybridEngine struct {
	yaegi ContractEngine
	gno   ContractEngine
}

func NewHybridEngine(yaegi ContractEngine, gno ContractEngine) ContractEngine {
	return hybridEngine{
		yaegi: yaegi,
		gno:   gno,
	}
}

func (e hybridEngine) ValidateContract(sourceCode string) base.OperationProcessReasonError {
	schema, err := AnalyzeContractSchema(sourceCode)
	if err != nil {
		return base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	switch schema.Mode {
	case SchemaModeTypedArgs:
		return e.gno.ValidateContract(sourceCode)
	default:
		return e.yaegi.ValidateContract(sourceCode)
	}
}

func (e hybridEngine) ExecuteContract(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req ExecuteRequest,
) (ExecuteResult, base.OperationProcessReasonError) {
	st, found, err := getStateFunc(pstate.RuntimeStateKey(req.Contract))
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	}

	if found {
		rv, err := pstate.GetRuntimeFromState(st)
		if err != nil {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
		}

		switch rv.Engine {
		case pstate.RuntimeEngineGnoSnapshot:
			return e.gno.ExecuteContract(encs, getStateFunc, req)
		case pstate.RuntimeEngineYaegi:
			return e.yaegi.ExecuteContract(encs, getStateFunc, req)
		default:
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError("unknown runtime engine %q", rv.Engine)
		}
	}

	schema, err := AnalyzeContractSchema(req.ContractCode)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	if schema.Mode == SchemaModeTypedArgs {
		return e.gno.ExecuteContract(encs, getStateFunc, req)
	}

	return e.yaegi.ExecuteContract(encs, getStateFunc, req)
}
func (e hybridEngine) QueryContract(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req QueryRequest,
) (QueryResult, base.OperationProcessReasonError) {
	st, found, err := getStateFunc(pstate.RuntimeStateKey(req.Contract))
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	}

	if found {
		rv, err := pstate.GetRuntimeFromState(st)
		if err != nil {
			return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
		}

		switch rv.Engine {
		case pstate.RuntimeEngineGnoSnapshot:
			return e.gno.QueryContract(encs, getStateFunc, req)
		case pstate.RuntimeEngineYaegi:
			return e.yaegi.QueryContract(encs, getStateFunc, req)
		default:
			return QueryResult{}, base.NewBaseOperationProcessReasonError("unknown runtime engine %q", rv.Engine)
		}
	}

	schema, err := AnalyzeContractSchema(req.ContractCode)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	if schema.Mode == SchemaModeTypedArgs {
		return e.gno.QueryContract(encs, getStateFunc, req)
	}

	return e.yaegi.QueryContract(encs, getStateFunc, req)
}
