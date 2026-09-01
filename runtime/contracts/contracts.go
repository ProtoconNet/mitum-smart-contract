package contracts

import ccontracts "github.com/imfact-labs/currency-model/app/runtime/contracts"

type ProposalOperationFactHintFunc = ccontracts.ProposalOperationFactHintFunc
type NewOperationProcessorInternalWithProposalFunc = ccontracts.NewOperationProcessorInternalWithProposalFunc

var (
	ProposalOperationFactHintContextKey = ccontracts.ProposalOperationFactHintContextKey
	OperationProcessorContextKey        = ccontracts.OperationProcessorContextKey
	OperationProcessorsMapBContextKey   = ccontracts.OperationProcessorsMapBContextKey
)
