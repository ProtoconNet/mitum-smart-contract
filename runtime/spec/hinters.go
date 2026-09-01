package spec

import (
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/smart-contract-model/operation/contract"
	"github.com/imfact-labs/smart-contract-model/state"
	"github.com/imfact-labs/smart-contract-model/types"
)

var AddedHinters = []encoder.DecodeDetail{
	{Hint: contract.RegisterContractHint, Instance: contract.RegisterContract{}},
	{Hint: contract.CallContractHint, Instance: contract.CallContract{}},
	{Hint: contract.CallContractItemHint, Instance: contract.CallContractItem{}},
	{Hint: types.DesignHint, Instance: types.Design{}},
	{Hint: state.DesignStateValueHint, Instance: state.DesignStateValue{}},
	{Hint: state.RuntimeStateValueHint, Instance: state.RuntimeStateValue{}},
	{Hint: state.SnapshotStateValueHint, Instance: state.SnapshotStateValue{}},
}

var AddedSupportedHinters = []encoder.DecodeDetail{
	{Hint: contract.RegisterContractFactHint, Instance: contract.RegisterContractFact{}},
	{Hint: contract.CallContractFactHint, Instance: contract.CallContractFact{}},
}
