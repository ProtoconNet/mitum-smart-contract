package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	CurrencyPolicyUpdaterFactHint = hint.MustNewHint("mitum-currency-currency-policy-updater-operation-fact-v0.0.1")
	CurrencyPolicyUpdaterHint     = hint.MustNewHint("mitum-currency-currency-policy-updater-operation-v0.0.1")
)

type CurrencyPolicyUpdaterFact struct {
	base.BaseFact
	currency CurrencyID
	policy   CurrencyPolicy
}

func NewCurrencyPolicyUpdaterFact(token []byte, currency CurrencyID, policy CurrencyPolicy) CurrencyPolicyUpdaterFact {
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
	if err := IsValidOperationFact(fact, b); err != nil {
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

func (fact CurrencyPolicyUpdaterFact) Currency() CurrencyID {
	return fact.currency
}

func (fact CurrencyPolicyUpdaterFact) Policy() CurrencyPolicy {
	return fact.policy
}

type CurrencyPolicyUpdater struct {
	BaseNodeOperation
}

func NewCurrencyPolicyUpdater(fact CurrencyPolicyUpdaterFact, memo string) (CurrencyPolicyUpdater, error) {
	return CurrencyPolicyUpdater{
		BaseNodeOperation: NewBaseNodeOperation(CurrencyPolicyUpdaterHint, fact),
	}, nil
}
