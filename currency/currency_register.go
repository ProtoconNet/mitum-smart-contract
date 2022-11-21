package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	CurrencyRegisterFactHint = hint.MustNewHint("mitum-currency-currency-register-operation-fact-v0.0.1")
	CurrencyRegisterHint     = hint.MustNewHint("mitum-currency-currency-register-operation-v0.0.1")
)

type CurrencyRegisterFact struct {
	base.BaseFact
	currency CurrencyDesign
}

func NewCurrencyRegisterFact(token []byte, de CurrencyDesign) CurrencyRegisterFact {
	fact := CurrencyRegisterFact{
		BaseFact: base.NewBaseFact(CurrencyRegisterHint, token),
		currency: de,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CurrencyRegisterFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CurrencyRegisterFact) Bytes() []byte {
	return util.ConcatBytesSlice(fact.Token(), fact.currency.Bytes())
}

func (fact CurrencyRegisterFact) IsValid(b []byte) error {
	if err := IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fact.currency); err != nil {
		return util.ErrInvalid.Errorf("invalid fact: %w", err)
	}

	if fact.currency.GenesisAccount() == nil {
		return util.ErrInvalid.Errorf("empty genesis account")
	}

	return nil
}

func (fact CurrencyRegisterFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CurrencyRegisterFact) Token() base.Token {
	return fact.Token()
}

func (fact CurrencyRegisterFact) Currency() CurrencyDesign {
	return fact.currency
}

type CurrencyRegister struct {
	base.BaseNodeOperation
}

func NewCurrencyRegister(fact CurrencyRegisterFact) (CurrencyRegister, error) {
	return CurrencyRegister{
		BaseNodeOperation: base.NewBaseNodeOperation(CurrencyRegisterHint, fact),
	}, nil
}
