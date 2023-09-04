package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	UpdateCurrencyFactHint = hint.MustNewHint("mitum-currency-update-currency-operation-fact-v0.0.1")
	UpdateCurrencyHint     = hint.MustNewHint("mitum-currency-update-currency-operation-v0.0.1")
)

type UpdateCurrencyFact struct {
	base.BaseFact
	currency types.CurrencyID
	policy   types.CurrencyPolicy
}

func NewUpdateCurrencyFact(token []byte, currency types.CurrencyID, policy types.CurrencyPolicy) UpdateCurrencyFact {
	fact := UpdateCurrencyFact{
		BaseFact: base.NewBaseFact(UpdateCurrencyFactHint, token),
		currency: currency,
		policy:   policy,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact UpdateCurrencyFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact UpdateCurrencyFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.currency.Bytes(),
		fact.policy.Bytes(),
	)
}

func (fact UpdateCurrencyFact) IsValid(b []byte) error {
	if err := common.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fact.currency, fact.policy); err != nil {
		return util.ErrInvalid.Errorf("invalid fact: %v", err)
	}

	return nil
}

func (fact UpdateCurrencyFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact UpdateCurrencyFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact UpdateCurrencyFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact UpdateCurrencyFact) Policy() types.CurrencyPolicy {
	return fact.policy
}

type UpdateCurrency struct {
	common.BaseNodeOperation
}

func NewUpdateCurrency(fact UpdateCurrencyFact, memo string) (UpdateCurrency, error) {
	return UpdateCurrency{
		BaseNodeOperation: common.NewBaseNodeOperation(UpdateCurrencyHint, fact),
	}, nil
}
