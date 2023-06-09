package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	CreateAccountsItemSingleAmountHint = hint.MustNewHint("mitum-currency-create-accounts-single-amount-v0.0.1")
)

type CreateAccountsItemSingleAmount struct {
	BaseCreateAccountsItem
}

func NewCreateAccountsItemSingleAmount(keys types.AccountKeys, amount types.Amount, addrType hint.Type) CreateAccountsItemSingleAmount {
	return CreateAccountsItemSingleAmount{
		BaseCreateAccountsItem: NewBaseCreateAccountsItem(CreateAccountsItemSingleAmountHint, keys, []types.Amount{amount}, addrType),
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
