package currency

import (
	base3 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	KeyUpdaterFactHint = hint.MustNewHint("mitum-currency-keyupdater-operation-fact-v0.0.1")
	KeyUpdaterHint     = hint.MustNewHint("mitum-currency-keyupdater-operation-v0.0.1")
)

type KeyUpdaterFact struct {
	base.BaseFact
	target   base.Address
	keys     base3.AccountKeys
	currency base3.CurrencyID
}

func NewKeyUpdaterFact(
	token []byte,
	target base.Address,
	keys base3.AccountKeys,
	currency base3.CurrencyID,
) KeyUpdaterFact {
	bf := base.NewBaseFact(KeyUpdaterFactHint, token)
	fact := KeyUpdaterFact{
		BaseFact: bf,
		target:   target,
		keys:     keys,
		currency: currency,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact KeyUpdaterFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact KeyUpdaterFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact KeyUpdaterFact) Bytes() []byte {
	return util.ConcatBytesSlice(
		fact.Token(),
		fact.target.Bytes(),
		fact.keys.Bytes(),
		fact.currency.Bytes(),
	)
}

func (fact KeyUpdaterFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := base3.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, fact.target, fact.keys, fact.currency); err != nil {
		return err
	}

	return nil
}

func (fact KeyUpdaterFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact KeyUpdaterFact) Target() base.Address {
	return fact.target
}

func (fact KeyUpdaterFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, 1)
	as[0] = fact.Target()
	return as, nil
}

func (fact KeyUpdaterFact) Keys() base3.AccountKeys {
	return fact.keys
}

func (fact KeyUpdaterFact) Currency() base3.CurrencyID {
	return fact.currency
}

func (fact KeyUpdaterFact) Rebuild() KeyUpdaterFact {
	fact.SetHash(fact.Hash())
	return fact
}

type KeyUpdater struct {
	base3.BaseOperation
}

func NewKeyUpdater(fact KeyUpdaterFact) (KeyUpdater, error) {
	return KeyUpdater{BaseOperation: base3.NewBaseOperation(KeyUpdaterHint, fact)}, nil
}

func (op *KeyUpdater) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
