package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	TransferItemMultiAmountsHint = hint.MustNewHint("mitum-currency-transfer-item-multi-amounts-v0.0.1")
)

var maxCurenciesTransferItemMultiAmounts = 10

type TransferItemMultiAmounts struct {
	BaseTransferItem
}

func NewTransferItemMultiAmounts(receiver base.Address, amounts []types.Amount) TransferItemMultiAmounts {
	return TransferItemMultiAmounts{
		BaseTransferItem: NewBaseTransferItem(TransferItemMultiAmountsHint, receiver, amounts),
	}
}

func (it TransferItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseTransferItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesTransferItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurenciesTransferItemMultiAmounts)
	}

	return nil
}

func (it TransferItemMultiAmounts) Rebuild() TransferItem {
	it.BaseTransferItem = it.BaseTransferItem.Rebuild().(BaseTransferItem)

	return it
}
