package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	KeyUpdaterFactHint = hint.MustNewHint("mitum-currency-keyupdater-operation-fact-v0.0.1")
	KeyUpdaterHint     = hint.MustNewHint("mitum-currency-keyupdater-operation-v0.0.1")
)

type KeyUpdaterFact struct {
	base.BaseFact
	target   base.Address
	keys     AccountKeys
	currency CurrencyID
}

func NewKeyUpdaterFact(
	token []byte,
	target base.Address,
	keys AccountKeys,
	currency CurrencyID,
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

	if err := IsValidOperationFact(fact, b); err != nil {
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

func (fact KeyUpdaterFact) Keys() AccountKeys {
	return fact.keys
}

func (fact KeyUpdaterFact) Rebuild() KeyUpdaterFact {
	fact.SetHash(fact.Hash())
	return fact
}

type KeyUpdater struct {
	BaseOperation
}

func NewKeyUpdater(fact KeyUpdaterFact, memo string) (KeyUpdater, error) {
	return KeyUpdater{BaseOperation: NewBaseOperationFromFact(KeyUpdaterHint, fact, memo)}, nil
}

func (op *KeyUpdater) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
