package currency

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var maxCurenciesCreateAccountsItemMultiAmounts = 10

var (
	CreateAccountsItemMultiAmountsHint = hint.MustNewHint("mitum-currency-create-accounts-multiple-amounts-v0.0.1")
)

type CreateAccountsItemMultiAmounts struct {
	BaseCreateAccountsItem
}

func NewCreateAccountsItemMultiAmounts(keys AccountKeys, amounts []Amount, addrType hint.Type) CreateAccountsItemMultiAmounts {
	return CreateAccountsItemMultiAmounts{
		BaseCreateAccountsItem: NewBaseCreateAccountsItem(CreateAccountsItemMultiAmountsHint, keys, amounts, addrType),
	}
}

func (it CreateAccountsItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseCreateAccountsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesCreateAccountsItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurenciesCreateAccountsItemMultiAmounts)
	}

	return nil
}

func (it CreateAccountsItemMultiAmounts) Rebuild() CreateAccountsItem {
	it.BaseCreateAccountsItem = it.BaseCreateAccountsItem.Rebuild().(BaseCreateAccountsItem)

	return it
}
