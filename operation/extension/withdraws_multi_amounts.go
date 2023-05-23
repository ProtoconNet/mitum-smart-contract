package extension

import (
	"github.com/ProtoconNet/mitum-currency/v2/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	WithdrawsItemMultiAmountsHint = hint.MustNewHint("mitum-currency-contract-account-withdraws-multi-amounts-v0.0.1")
)

var maxCurenciesWithdrawsItemMultiAmounts = 10

type WithdrawsItemMultiAmounts struct {
	BaseWithdrawsItem
}

func NewWithdrawsItemMultiAmounts(target mitumbase.Address, amounts []base.Amount) WithdrawsItemMultiAmounts {
	return WithdrawsItemMultiAmounts{
		BaseWithdrawsItem: NewBaseWithdrawsItem(WithdrawsItemMultiAmountsHint, target, amounts),
	}
}

func (it WithdrawsItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseWithdrawsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesWithdrawsItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurenciesWithdrawsItemMultiAmounts)
	}

	return nil
}

func (it WithdrawsItemMultiAmounts) Rebuild() WithdrawsItem {
	it.BaseWithdrawsItem = it.BaseWithdrawsItem.Rebuild().(BaseWithdrawsItem)

	return it
}
