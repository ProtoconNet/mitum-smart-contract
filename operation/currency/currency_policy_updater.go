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
	CurrencyPolicyUpdaterFactHint = hint.MustNewHint("mitum-currency-currency-policy-updater-operation-fact-v0.0.1")
	CurrencyPolicyUpdaterHint     = hint.MustNewHint("mitum-currency-currency-policy-updater-operation-v0.0.1")
)

type CurrencyPolicyUpdaterFact struct {
	base.BaseFact
	currency types.CurrencyID
	policy   types.CurrencyPolicy
}

func NewCurrencyPolicyUpdaterFact(token []byte, currency types.CurrencyID, policy types.CurrencyPolicy) CurrencyPolicyUpdaterFact {
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
	if err := common.IsValidOperationFact(fact, b); err != nil {
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

func (fact CurrencyPolicyUpdaterFact) Currency() types.CurrencyID {
	return fact.currency
}

func (fact CurrencyPolicyUpdaterFact) Policy() types.CurrencyPolicy {
	return fact.policy
}

type CurrencyPolicyUpdater struct {
	common.BaseNodeOperation
}

func NewCurrencyPolicyUpdater(fact CurrencyPolicyUpdaterFact, memo string) (CurrencyPolicyUpdater, error) {
	return CurrencyPolicyUpdater{
		BaseNodeOperation: common.NewBaseNodeOperation(CurrencyPolicyUpdaterHint, fact),
	}, nil
}
