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
	CurrencyRegisterFactHint = hint.MustNewHint("mitum-currency-currency-register-operation-fact-v0.0.1")
	CurrencyRegisterHint     = hint.MustNewHint("mitum-currency-currency-register-operation-v0.0.1")
)

type CurrencyRegisterFact struct {
	base.BaseFact
	currency types.CurrencyDesign
}

func NewCurrencyRegisterFact(token []byte, de types.CurrencyDesign) CurrencyRegisterFact {
	fact := CurrencyRegisterFact{
		BaseFact: base.NewBaseFact(CurrencyRegisterFactHint, token),
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
	if err := common.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fact.currency); err != nil {
		return util.ErrInvalid.Errorf("invalid fact: %v", err)
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
	return fact.BaseFact.Token()
}

func (fact CurrencyRegisterFact) Currency() types.CurrencyDesign {
	return fact.currency
}

type CurrencyRegister struct {
	common.BaseNodeOperation
}

func NewCurrencyRegister(fact CurrencyRegisterFact, memo string) (CurrencyRegister, error) {
	return CurrencyRegister{
		BaseNodeOperation: common.NewBaseNodeOperation(CurrencyRegisterHint, fact),
	}, nil
}
