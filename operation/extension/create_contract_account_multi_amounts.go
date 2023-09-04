package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var maxCurrenciesCreateContractAccountItemMultiAmounts = 10

var (
	CreateContractAccountItemMultiAmountsHint = hint.MustNewHint("mitum-currency-create-contract-account-multiple-amounts-v0.0.1")
)

type CreateContractAccountItemMultiAmounts struct {
	BaseCreateContractAccountItem
}

func NewCreateContractAccountItemMultiAmounts(keys types.AccountKeys, amounts []types.Amount, addrType hint.Type) CreateContractAccountItemMultiAmounts {
	return CreateContractAccountItemMultiAmounts{
		BaseCreateContractAccountItem: NewBaseCreateContractAccountItem(CreateContractAccountItemMultiAmountsHint, keys, amounts, addrType),
	}
}

func (it CreateContractAccountItemMultiAmounts) IsValid([]byte) error {
	if err := it.BaseCreateContractAccountItem.IsValid(nil); err != nil {
		return err
	}

	if n := len(it.amounts); n > maxCurrenciesCreateContractAccountItemMultiAmounts {
		return util.ErrInvalid.Errorf("amounts over allowed; %d > %d", n, maxCurrenciesCreateContractAccountItemMultiAmounts)
	}

	return nil
}

func (it CreateContractAccountItemMultiAmounts) Rebuild() CreateContractAccountItem {
	it.BaseCreateContractAccountItem = it.BaseCreateContractAccountItem.Rebuild().(BaseCreateContractAccountItem)

	return it
}
