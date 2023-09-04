package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	CreateAccountItemSingleAmountHint = hint.MustNewHint("mitum-currency-create-account-single-amount-v0.0.1")
)

type CreateAccountItemSingleAmount struct {
	BaseCreateAccountItem
}

func NewCreateAccountItemSingleAmount(keys types.AccountKeys, amount types.Amount, addrType hint.Type) CreateAccountItemSingleAmount {
	return CreateAccountItemSingleAmount{
		BaseCreateAccountItem: NewBaseCreateAccountItem(CreateAccountItemSingleAmountHint, keys, []types.Amount{amount}, addrType),
	}
}

func (it CreateAccountItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseCreateAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return util.ErrInvalid.Errorf("only one amount allowed; %d", n)
	}

	return nil
}

func (it CreateAccountItemSingleAmount) Rebuild() CreateAccountItem {
	it.BaseCreateAccountItem = it.BaseCreateAccountItem.Rebuild().(BaseCreateAccountItem)

	return it
}
