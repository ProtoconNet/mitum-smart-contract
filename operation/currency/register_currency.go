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
	RegisterCurrencyFactHint = hint.MustNewHint("mitum-currency-register-currency-operation-fact-v0.0.1")
	RegisterCurrencyHint     = hint.MustNewHint("mitum-currency-register-currency-operation-v0.0.1")
)

type RegisterCurrencyFact struct {
	base.BaseFact
	currency types.CurrencyDesign
}

func NewRegisterCurrencyFact(token []byte, de types.CurrencyDesign) RegisterCurrencyFact {
	fact := RegisterCurrencyFact{
		BaseFact: base.NewBaseFact(RegisterCurrencyFactHint, token),
		currency: de,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact RegisterCurrencyFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact RegisterCurrencyFact) Bytes() []byte {
	return util.ConcatBytesSlice(fact.Token(), fact.currency.Bytes())
}

func (fact RegisterCurrencyFact) IsValid(b []byte) error {
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

func (fact RegisterCurrencyFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact RegisterCurrencyFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact RegisterCurrencyFact) Currency() types.CurrencyDesign {
	return fact.currency
}

type RegisterCurrency struct {
	common.BaseNodeOperation
}

func NewRegisterCurrency(fact RegisterCurrencyFact, memo string) (RegisterCurrency, error) {
	return RegisterCurrency{
		BaseNodeOperation: common.NewBaseNodeOperation(RegisterCurrencyHint, fact),
	}, nil
}
