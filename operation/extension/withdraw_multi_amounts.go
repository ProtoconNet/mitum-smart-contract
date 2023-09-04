package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	WithdrawItemMultiAmountsHint = hint.MustNewHint("mitum-currency-contract-account-withdraw-multi-amounts-v0.0.1")
)

var maxCurenciesWithdrawItemMultiAmounts = 10

type WithdrawItemMultiAmounts struct {
	BaseWithdrawItem
}

func NewWithdrawItemMultiAmounts(target base.Address, amounts []types.Amount) WithdrawItemMultiAmounts {
	return WithdrawItemMultiAmounts{
		BaseWithdrawItem: NewBaseWithdrawItem(WithdrawItemMultiAmountsHint, target, amounts),
	}
}

func (it WithdrawItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseWithdrawItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesWithdrawItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurenciesWithdrawItemMultiAmounts)
	}

	return nil
}

func (it WithdrawItemMultiAmounts) Rebuild() WithdrawItem {
	it.BaseWithdrawItem = it.BaseWithdrawItem.Rebuild().(BaseWithdrawItem)

	return it
}
