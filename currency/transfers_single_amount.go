package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

var (
	TransfersItemSingleAmountHint = hint.MustNewHint("mitum-currency-transfers-item-single-amount-v0.0.1")
)

type TransfersItemSingleAmount struct {
	BaseTransfersItem
}

func NewTransfersItemSingleAmount(receiver base.Address, amount Amount) TransfersItemSingleAmount {
	return TransfersItemSingleAmount{
		BaseTransfersItem: NewBaseTransfersItem(TransfersItemSingleAmountHint, receiver, []Amount{amount}),
	}
}

func (it TransfersItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseTransfersItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return util.ErrInvalid.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it TransfersItemSingleAmount) Rebuild() TransfersItem {
	it.BaseTransfersItem = it.BaseTransfersItem.Rebuild().(BaseTransfersItem)

	return it
}
