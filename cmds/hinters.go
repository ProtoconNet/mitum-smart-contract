package cmds

import (
	"github.com/imfact-labs/currency-model/app/runtime/spec"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util/encoder"
	runtimespec "github.com/imfact-labs/smart-contract-model/runtime/spec"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
)

var Hinters []encoder.DecodeDetail
var SupportedProposalOperationFactHinters []encoder.DecodeDetail

var AddedHinters = runtimespec.AddedHinters
var AddedSupportedHinters = runtimespec.AddedSupportedHinters

func init() {
	Hinters = append(Hinters, spec.Hinters...)
	Hinters = append(Hinters, AddedHinters...)

	SupportedProposalOperationFactHinters = append(
		SupportedProposalOperationFactHinters,
		launch.SupportedProposalOperationFactHinters...,
	)
	SupportedProposalOperationFactHinters = append(
		SupportedProposalOperationFactHinters, spec.AddedSupportedHinters...)
	SupportedProposalOperationFactHinters = append(
		SupportedProposalOperationFactHinters, AddedSupportedHinters...)
}

func LoadHinters(encs *encoder.Encoders) error {
	return steps.LoadHinters(encs)
}
