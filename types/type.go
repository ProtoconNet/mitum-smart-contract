package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
)

type GetNewProcessor func(
	height base.Height,
	getStateFunc base.GetStateFunc,
	newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc base.NewOperationProcessorProcessFunc) (base.OperationProcessor, error)

type GetNewProcessorWithProposal func(
	height base.Height,
	proposal *base.ProposalSignFact,
	getStateFunc base.GetStateFunc,
	newPreProcessConstraintFunc base.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc base.NewOperationProcessorProcessFunc) (base.OperationProcessor, error)

type DuplicationType string

type AddFee map[CurrencyID][2]common.Big

func (af AddFee) Fee(key CurrencyID, fee common.Big) AddFee {
	switch v, found := af[key]; {
	case !found:
		af[key] = [2]common.Big{common.ZeroBig, fee}
	default:
		af[key] = [2]common.Big{v[0], v[1].Add(fee)}
	}

	return af
}

func (af AddFee) Add(key CurrencyID, add common.Big) AddFee {
	switch v, found := af[key]; {
	case !found:
		af[key] = [2]common.Big{add, common.ZeroBig}
	default:
		af[key] = [2]common.Big{v[0].Add(add), v[1]}
	}

	return af
}
