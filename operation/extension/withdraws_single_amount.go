package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	WithdrawsItemSingleAmountHint = hint.MustNewHint("mitum-currency-contract-account-withdraws-single-amount-v0.0.1")
)

type WithdrawsItemSingleAmount struct {
	BaseWithdrawsItem
}

func NewWithdrawsItemSingleAmount(target base.Address, amount types.Amount) WithdrawsItemSingleAmount {
	return WithdrawsItemSingleAmount{
		BaseWithdrawsItem: NewBaseWithdrawsItem(WithdrawsItemSingleAmountHint, target, []types.Amount{amount}),
	}
}

func (it WithdrawsItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseWithdrawsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return util.ErrInvalid.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it WithdrawsItemSingleAmount) Rebuild() WithdrawsItem {
	it.BaseWithdrawsItem = it.BaseWithdrawsItem.Rebuild().(BaseWithdrawsItem)

	return it
}
