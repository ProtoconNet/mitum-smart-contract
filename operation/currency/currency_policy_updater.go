package currency

import (
	base3 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	CurrencyPolicyUpdaterFactHint = hint.MustNewHint("mitum-currency-currency-policy-updater-operation-fact-v0.0.1")
	CurrencyPolicyUpdaterHint     = hint.MustNewHint("mitum-currency-currency-policy-updater-operation-v0.0.1")
)

type CurrencyPolicyUpdaterFact struct {
	base.BaseFact
	currency base3.CurrencyID
	policy   base3.CurrencyPolicy
}

func NewCurrencyPolicyUpdaterFact(token []byte, currency base3.CurrencyID, policy base3.CurrencyPolicy) CurrencyPolicyUpdaterFact {
	fact := CurrencyPolicyUpdaterFact{
		BaseFact: base.NewBaseFact(CurrencyPolicyUpdaterFactHint, token),
		currency: currency,
		policy:   policy,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CurrencyPolicyUpdaterFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CurrencyPolicyUpdaterFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.currency.Bytes(),
		fact.policy.Bytes(),
	)
}

func (fact CurrencyPolicyUpdaterFact) IsValid(b []byte) error {
	if err := base3.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fact.currency, fact.policy); err != nil {
		return util.ErrInvalid.Errorf("invalid fact: %w", err)
	}

	return nil
}

func (fact CurrencyPolicyUpdaterFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CurrencyPolicyUpdaterFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact CurrencyPolicyUpdaterFact) Currency() base3.CurrencyID {
	return fact.currency
}

func (fact CurrencyPolicyUpdaterFact) Policy() base3.CurrencyPolicy {
	return fact.policy
}

type CurrencyPolicyUpdater struct {
	base3.BaseNodeOperation
}

func NewCurrencyPolicyUpdater(fact CurrencyPolicyUpdaterFact, memo string) (CurrencyPolicyUpdater, error) {
	return CurrencyPolicyUpdater{
		BaseNodeOperation: base3.NewBaseNodeOperation(CurrencyPolicyUpdaterHint, fact),
	}, nil
}
