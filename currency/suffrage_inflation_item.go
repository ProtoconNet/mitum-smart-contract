package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
)

type SuffrageInflationItem struct {
	receiver base.Address
	amount   Amount
}

func NewSuffrageInflationItem(receiver base.Address, amount Amount) SuffrageInflationItem {
	return SuffrageInflationItem{
		receiver: receiver,
		amount:   amount,
	}
}

func (it SuffrageInflationItem) Bytes() []byte {
	var br []byte
	if it.receiver != nil {
		br = it.receiver.Bytes()
	}

	return util.ConcatBytesSlice(br, it.amount.Bytes())
}

func (it SuffrageInflationItem) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, it.receiver, it.amount); err != nil {
		return err
	}

	if !it.amount.Big().OverZero() {
		return util.ErrInvalid.Errorf("under zero amount of SuffrageInflationItemo")
	}

	return nil
}
