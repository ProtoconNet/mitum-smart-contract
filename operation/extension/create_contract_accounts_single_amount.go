package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	CreateContractAccountsItemSingleAmountHint = hint.MustNewHint("mitum-currency-create-contract-accounts-single-amount-v0.0.1")
)

type CreateContractAccountsItemSingleAmount struct {
	BaseCreateContractAccountsItem
}

func NewCreateContractAccountsItemSingleAmount(keys types.AccountKeys, amount types.Amount, addrType hint.Type) CreateContractAccountsItemSingleAmount {
	return CreateContractAccountsItemSingleAmount{
		BaseCreateContractAccountsItem: NewBaseCreateContractAccountsItem(CreateContractAccountsItemSingleAmountHint, keys, []types.Amount{amount}, addrType),
	}
}

func (it CreateContractAccountsItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseCreateContractAccountsItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return util.ErrInvalid.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it CreateContractAccountsItemSingleAmount) Rebuild() CreateContractAccountsItem {
	it.BaseCreateContractAccountsItem = it.BaseCreateContractAccountsItem.Rebuild().(BaseCreateContractAccountsItem)

	return it
}
