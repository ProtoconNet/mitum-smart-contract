package cmds

import (
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
)

func IsSupportedProposalOperationFactHintFunc() func(hint.Hint) bool {
	return steps.IsSupportedProposalOperationFactHintFunc()
}
