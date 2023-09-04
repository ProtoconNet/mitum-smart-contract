package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	WithdrawItemSingleAmountHint = hint.MustNewHint("mitum-currency-contract-account-withdraw-single-amount-v0.0.1")
)

type WithdrawItemSingleAmount struct {
	BaseWithdrawItem
}

func NewWithdrawItemSingleAmount(target base.Address, amount types.Amount) WithdrawItemSingleAmount {
	return WithdrawItemSingleAmount{
		BaseWithdrawItem: NewBaseWithdrawItem(WithdrawItemSingleAmountHint, target, []types.Amount{amount}),
	}
}

func (it WithdrawItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseWithdrawItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return util.ErrInvalid.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it WithdrawItemSingleAmount) Rebuild() WithdrawItem {
	it.BaseWithdrawItem = it.BaseWithdrawItem.Rebuild().(BaseWithdrawItem)

	return it
}
