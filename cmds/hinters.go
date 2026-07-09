package cmds

import (
	"github.com/imfact-labs/currency-model/app/runtime/spec"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	pstate "github.com/imfact-labs/smart-contract-model/state"
	ptypes "github.com/imfact-labs/smart-contract-model/types"
	"github.com/pkg/errors"
)

var Hinters []encoder.DecodeDetail
var SupportedProposalOperationFactHinters []encoder.DecodeDetail

var AddedHinters = []encoder.DecodeDetail{
	// revive:disable-next-line:line-length-limit
	{Hint: contract.RegisterContractHint, Instance: contract.RegisterContract{}},
	{Hint: contract.CallContractHint, Instance: contract.CallContract{}},
	{Hint: contract.CallContractItemHint, Instance: contract.CallContractItem{}},
	{Hint: ptypes.DesignHint, Instance: ptypes.Design{}},
	{Hint: pstate.DesignStateValueHint, Instance: pstate.DesignStateValue{}},
	{Hint: pstate.RuntimeStateValueHint, Instance: pstate.RuntimeStateValue{}},
	{Hint: pstate.SnapshotStateValueHint, Instance: pstate.SnapshotStateValue{}},
}

var AddedSupportedHinters = []encoder.DecodeDetail{
	{Hint: contract.RegisterContractFactHint, Instance: contract.RegisterContractFact{}},
	{Hint: contract.CallContractFactHint, Instance: contract.CallContractFact{}},
}

func init() {
	defaultLen := len(launch.Hinters)
	currencyExtendedLen := defaultLen + len(spec.AddedHinters)
	allExtendedLen := currencyExtendedLen + len(AddedHinters)

	Hinters = make([]encoder.DecodeDetail, allExtendedLen)
	copy(Hinters, launch.Hinters)
	copy(Hinters[defaultLen:currencyExtendedLen], spec.AddedHinters)
	copy(Hinters[currencyExtendedLen:], AddedHinters)

	defaultSupportedLen := len(launch.SupportedProposalOperationFactHinters)
	currencySupportedExtendedLen := defaultSupportedLen + len(spec.AddedSupportedHinters)
	allSupportedExtendedLen := currencySupportedExtendedLen + len(AddedSupportedHinters)

	SupportedProposalOperationFactHinters = make(
		[]encoder.DecodeDetail,
		allSupportedExtendedLen)
	copy(SupportedProposalOperationFactHinters, launch.SupportedProposalOperationFactHinters)
	copy(SupportedProposalOperationFactHinters[defaultSupportedLen:currencySupportedExtendedLen], spec.AddedSupportedHinters)
	copy(SupportedProposalOperationFactHinters[currencySupportedExtendedLen:], AddedSupportedHinters)
}

func LoadHinters(encs *encoder.Encoders) error {
	for i := range AddedHinters {
		if err := encs.AddDetail(AddedHinters[i]); err != nil {
			return errors.Wrap(err, "add hinter to encoder")
		}
	}

	for i := range AddedSupportedHinters {
		if err := encs.AddDetail(AddedSupportedHinters[i]); err != nil {
			return errors.Wrap(err, "add supported proposal operation fact hinter to encoder")
		}
	}

	return nil
}
