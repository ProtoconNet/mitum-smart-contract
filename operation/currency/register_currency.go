package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
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
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, fact.currency); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	if fact.currency.GenesisAccount() == nil {
		return common.ErrFactInvalid.Wrap(common.ErrValOOR.Wrap(errors.Errorf("Value out of range: Empty genesis account")))
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
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

func NewRegisterCurrency(fact RegisterCurrencyFact) (RegisterCurrency, error) {
	return RegisterCurrency{
		BaseNodeOperation: common.NewBaseNodeOperation(RegisterCurrencyHint, fact),
	}, nil
}
