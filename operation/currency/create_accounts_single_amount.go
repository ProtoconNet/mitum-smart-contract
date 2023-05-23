package currency

import (
	base2 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	CreateAccountsItemSingleAmountHint = hint.MustNewHint("mitum-currency-create-accounts-single-amount-v0.0.1")
)

type CreateAccountsItemSingleAmount struct {
	BaseCreateAccountsItem
}

func NewCreateAccountsItemSingleAmount(keys base2.AccountKeys, amount base2.Amount, addrType hint.Type) CreateAccountsItemSingleAmount {
	return CreateAccountsItemSingleAmount{
		BaseCreateAccountsItem: NewBaseCreateAccountsItem(CreateAccountsItemSingleAmountHint, keys, []base2.Amount{amount}, addrType),
	}
}

func (it CreateAccountsItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseCreateAccountsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return util.ErrInvalid.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it CreateAccountsItemSingleAmount) Rebuild() CreateAccountsItem {
	it.BaseCreateAccountsItem = it.BaseCreateAccountsItem.Rebuild().(BaseCreateAccountsItem)

	return it
}
