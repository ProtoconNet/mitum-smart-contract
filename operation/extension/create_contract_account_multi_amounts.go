package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var maxCurrenciesCreateContractAccountItemMultiAmounts = 10

var (
	CreateContractAccountItemMultiAmountsHint = hint.MustNewHint("mitum-extension-create-contract-account-multiple-amounts-v0.0.1")
)

type CreateContractAccountItemMultiAmounts struct {
	BaseCreateContractAccountItem
}

func NewCreateContractAccountItemMultiAmounts(keys types.AccountKeys, amounts []types.Amount) CreateContractAccountItemMultiAmounts {
	return CreateContractAccountItemMultiAmounts{
		BaseCreateContractAccountItem: NewBaseCreateContractAccountItem(CreateContractAccountItemMultiAmountsHint, keys, amounts),
	}
}

func (it CreateContractAccountItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseCreateContractAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurrenciesCreateContractAccountItemMultiAmounts {
		return common.ErrArrayLen.Wrap(errors.Errorf("amounts over allowed; %d > %d", n, maxCurrenciesCreateContractAccountItemMultiAmounts))
	}

	return nil
}

func (it CreateContractAccountItemMultiAmounts) Rebuild() CreateContractAccountItem {
	it.BaseCreateContractAccountItem = it.BaseCreateContractAccountItem.Rebuild().(BaseCreateContractAccountItem)

	return it
}
