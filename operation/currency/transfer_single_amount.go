package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	TransferItemSingleAmountHint = hint.MustNewHint("mitum-currency-transfer-item-single-amount-v0.0.1")
)

type TransferItemSingleAmount struct {
	BaseTransferItem
}

func NewTransferItemSingleAmount(receiver base.Address, amount types.Amount) TransferItemSingleAmount {
	return TransferItemSingleAmount{
		BaseTransferItem: NewBaseTransferItem(TransferItemSingleAmountHint, receiver, []types.Amount{amount}),
	}
}

func (it TransferItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseTransferItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return common.ErrArrayLen.Wrap(errors.Errorf("only one amount allowed, %d", n))
	}

	return nil
}

func (it TransferItemSingleAmount) Rebuild() TransferItem {
	it.BaseTransferItem = it.BaseTransferItem.Rebuild().(BaseTransferItem)

	return it
}
