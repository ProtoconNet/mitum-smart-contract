package types

import (
	"github.com/ProtoconNet/mitum-currency/v2/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
)

type GetNewProcessor func(
	height mitumbase.Height,
	getStateFunc mitumbase.GetStateFunc,
	newPreProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc,
	newProcessConstraintFunc mitumbase.NewOperationProcessorProcessFunc) (mitumbase.OperationProcessor, error)

type DuplicationType string

type AddFee map[base.CurrencyID][2]base.Big

func (af AddFee) Fee(key base.CurrencyID, fee base.Big) AddFee {
	switch v, found := af[key]; {
	case !found:
		af[key] = [2]base.Big{base.ZeroBig, fee}
	default:
		af[key] = [2]base.Big{v[0], v[1].Add(fee)}
	}

	return af
}

func (af AddFee) Add(key base.CurrencyID, add base.Big) AddFee {
	switch v, found := af[key]; {
	case !found:
		af[key] = [2]base.Big{add, base.ZeroBig}
	default:
		af[key] = [2]base.Big{v[0].Add(add), v[1]}
	}

	return af
}
