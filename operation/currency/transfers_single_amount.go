package currency

import (
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	TransfersItemSingleAmountHint = hint.MustNewHint("mitum-currency-transfers-item-single-amount-v0.0.1")
)

type TransfersItemSingleAmount struct {
	BaseTransfersItem
}

func NewTransfersItemSingleAmount(receiver base.Address, amount base2.Amount) TransfersItemSingleAmount {
	return TransfersItemSingleAmount{
		BaseTransfersItem: NewBaseTransfersItem(TransfersItemSingleAmountHint, receiver, []base2.Amount{amount}),
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
