package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var maxCurenciesCreateAccountItemMultiAmounts = 10

var (
	CreateAccountItemMultiAmountsHint = hint.MustNewHint("mitum-currency-create-account-multiple-amounts-v0.0.1")
)

type CreateAccountItemMultiAmounts struct {
	BaseCreateAccountItem
}

func NewCreateAccountItemMultiAmounts(keys types.AccountKeys, amounts []types.Amount) CreateAccountItemMultiAmounts {
	return CreateAccountItemMultiAmounts{
		BaseCreateAccountItem: NewBaseCreateAccountItem(CreateAccountItemMultiAmountsHint, keys, amounts),
	}
}

func (it CreateAccountItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseCreateAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurenciesCreateAccountItemMultiAmounts {
		return common.ErrValOOR.Wrap(errors.Errorf("Amounts over allowed, %d > %d", n, maxCurenciesCreateAccountItemMultiAmounts))
	}

	return nil
}

func (it CreateAccountItemMultiAmounts) Rebuild() CreateAccountItem {
	it.BaseCreateAccountItem = it.BaseCreateAccountItem.Rebuild().(BaseCreateAccountItem)

	return it
}
