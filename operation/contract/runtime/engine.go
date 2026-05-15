package runtime

import (
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type InvocationMode string

const (
	InvocationModeRegister InvocationMode = "register"
	InvocationModeCall     InvocationMode = "call"
)

type ExecuteRequest struct {
	Mode         InvocationMode
	Contract     base.Address
	Sender       base.Address
	Height       base.Height
	ContractCode string
	Function     string
	CallData     map[string]string
}

type ExecuteResult struct {
	Engine      pstate.RuntimeEngine
	Data        map[string]interface{}
	StateMerges []base.StateMergeValue
}

type QueryRequest struct {
	Contract     base.Address
	Sender       base.Address
	Height       base.Height
	ContractCode string
	Function     string
	CallData     map[string]string
}

type QueryResult struct {
	Engine pstate.RuntimeEngine
	Result interface{}
	Ok     *bool
}

type ContractEngine interface {
	ValidateContract(sourceCode string) base.OperationProcessReasonError
	ExecuteContract(
		encs encoder.Encoders,
		getStateFunc base.GetStateFunc,
		req ExecuteRequest,
	) (ExecuteResult, base.OperationProcessReasonError)
	QueryContract(
		encs encoder.Encoders,
		getStateFunc base.GetStateFunc,
		req QueryRequest,
	) (QueryResult, base.OperationProcessReasonError)
}
