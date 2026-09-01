package cmds

import (
	"context"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
)

var PNameOperationProcessorsMap = steps.PNameOperationProcessorsMap

type NewOperationProcessorInternalWithProposalFunc func(base.Height, base.ProposalSignFact, base.GetStateFunc) (base.OperationProcessor, error)

func POperationProcessorsMap(pctx context.Context) (context.Context, error) {
	return steps.POperationProcessorsMap(pctx)
}
