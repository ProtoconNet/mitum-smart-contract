package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var maxCurenciesCreateAccountItemMultiAmounts = 10

var (
	CreateAccountItemMultiAmountsHint = hint.MustNewHint("mitum-currency-create-account-multiple-amounts-v0.0.1")
)

type CreateAccountItemMultiAmounts struct {
	BaseCreateAccountItem
}

func NewCreateAccountItemMultiAmounts(keys types.AccountKeys, amounts []types.Amount, addrType hint.Type) CreateAccountItemMultiAmounts {
	return CreateAccountItemMultiAmounts{
		BaseCreateAccountItem: NewBaseCreateAccountItem(CreateAccountItemMultiAmountsHint, keys, amounts, addrType),
	}
}

func (it CreateAccountItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseCreateAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesCreateAccountItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurenciesCreateAccountItemMultiAmounts)
	}

	return nil
}

func (it CreateAccountItemMultiAmounts) Rebuild() CreateAccountItem {
	it.BaseCreateAccountItem = it.BaseCreateAccountItem.Rebuild().(BaseCreateAccountItem)

	return it
}
