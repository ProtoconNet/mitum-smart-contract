package processor

import (
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
)

func ExecutionClass(op base.Operation) isaac.OperationExecutionClass {
	switch op.Hint().Type() {
	case contract.RegisterContractHint.Type(),
		contract.CallContractHint.Type():
		return isaac.OperationExecutionClassSerial
	default:
		return isaac.OperationExecutionClassParallel
	}
}
