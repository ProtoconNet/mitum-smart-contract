package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	CreateContractAccountItemSingleAmountHint = hint.MustNewHint("mitum-extension-create-contract-account-single-amount-v0.0.1")
)

type CreateContractAccountItemSingleAmount struct {
	BaseCreateContractAccountItem
}

func NewCreateContractAccountItemSingleAmount(keys types.AccountKeys, amount types.Amount) CreateContractAccountItemSingleAmount {
	return CreateContractAccountItemSingleAmount{
		BaseCreateContractAccountItem: NewBaseCreateContractAccountItem(CreateContractAccountItemSingleAmountHint, keys, []types.Amount{amount}),
	}
}

func (it CreateContractAccountItemSingleAmount) IsValid([]byte) error {
	if err := it.BaseCreateContractAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n != 1 {
		return common.ErrArrayLen.Wrap(errors.Errorf("only one amount allowed, %d", n))
	}

	return nil
}

func (it CreateContractAccountItemSingleAmount) Rebuild() CreateContractAccountItem {
	it.BaseCreateContractAccountItem = it.BaseCreateContractAccountItem.Rebuild().(BaseCreateContractAccountItem)

	return it
}
