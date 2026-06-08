package runtime

import (
	"github.com/ProtoconNet/mitum-smart-contract/state"
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
	BlockTime    int64
	ContractCode string
	Schema       *ContractSchema

	// Function and CallData are kept as a legacy single-call bridge. New call
	// execution should normalize through CallItems; register should use InitData.
	Function  string
	CallData  map[string]string
	CallItems []ExecuteCallItem
	InitData  map[string]string
}

type ExecuteCallItem struct {
	Function string
	CallData map[string]string
}

type ExecuteResult struct {
	Engine      state.RuntimeEngine
	StateMerges []base.StateMergeValue
}

type QueryRequest struct {
	Contract      base.Address
	Sender        base.Address
	Height        base.Height
	CurrentHeight base.Height
	ContractCode  string
	Schema        *ContractSchema
	Function      string
	CallData      map[string]string
}

type QueryResult struct {
	Engine state.RuntimeEngine
	Result interface{}
	Ok     *bool
}

type ContractEngine interface {
	ValidateContract(sourceCode string) (ContractSchema, base.OperationProcessReasonError)
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
