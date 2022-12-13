package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

var (
	TransfersItemMultiAmountsHint = hint.MustNewHint("mitum-currency-transfers-item-multi-amounts-v0.0.1")
)

var maxCurenciesTransfersItemMultiAmounts = 10

type TransfersItemMultiAmounts struct {
	BaseTransfersItem
}

func NewTransfersItemMultiAmounts(receiver base.Address, amounts []Amount) TransfersItemMultiAmounts {
	return TransfersItemMultiAmounts{
		BaseTransfersItem: NewBaseTransfersItem(TransfersItemMultiAmountsHint, receiver, amounts),
	}
}

func (it TransfersItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseTransfersItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesTransfersItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurenciesTransfersItemMultiAmounts)
	}

	return nil
}

func (it TransfersItemMultiAmounts) Rebuild() TransfersItem {
	it.BaseTransfersItem = it.BaseTransfersItem.Rebuild().(BaseTransfersItem)

	return it
}
