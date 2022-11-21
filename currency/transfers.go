package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	TransfersFactHint = hint.MustNewHint("mitum-currency-transfers-operation-fact-v0.0.1")
	TransfersHint     = hint.MustNewHint("mitum-currency-transfers-operation-v0.0.1")
)

var MaxTransferItems uint = 10

type TransfersItem interface {
	hint.Hinter
	util.IsValider
	AmountsItem
	Bytes() []byte
	Receiver() base.Address
	Rebuild() TransfersItem
}

type TransfersFact struct {
	base.BaseFact
	sender base.Address
	items  []TransfersItem
}

func NewTransfersFact(
	token []byte,
	sender base.Address,
	items []TransfersItem,
) TransfersFact {
	bf := base.NewBaseFact(TransfersFactHint, token)
	fact := TransfersFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact TransfersFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact TransfersFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact TransfersFact) Bytes() []byte {
	its := make([][]byte, len(fact.items))
	for i := range fact.items {
		its[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		util.ConcatBytesSlice(its...),
	)
}

func (fact TransfersFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.items); n < 1 {
		return util.ErrInvalid.Errorf("empty items")
	} else if n > int(MaxTransferItems) {
		return util.ErrInvalid.Errorf("items, %d over max, %d", n, MaxTransferItems)
	}

	if err := util.CheckIsValiders(nil, false, fact.sender); err != nil {
		return err
	}

	foundReceivers := map[string]struct{}{}
	for i := range fact.items {
		it := fact.items[i]
		if err := util.CheckIsValiders(nil, false, it); err != nil {
			return err
		}

		k := it.Receiver().String()
		switch _, found := foundReceivers[k]; {
		case found:
			return util.ErrInvalid.Errorf("duplicated receiver found, %s", it.Receiver())
		case fact.sender.Equal(it.Receiver()):
			return util.ErrInvalid.Errorf("receiver is same with sender, %q", fact.sender)
		default:
			foundReceivers[k] = struct{}{}
		}
	}

	return nil
}

func (fact TransfersFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact TransfersFact) Sender() base.Address {
	return fact.sender
}

func (fact TransfersFact) Items() []TransfersItem {
	return fact.items
}

func (fact TransfersFact) Rebuild() TransfersFact {
	items := make([]TransfersItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.SetHash(fact.Hash())

	return fact
}

func (fact TransfersFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)
	for i := range fact.items {
		as[i] = fact.items[i].Receiver()
	}

	as[len(fact.items)] = fact.Sender()

	return as, nil
}

type Transfers struct {
	BaseOperation
}

func NewTransfers(fact TransfersFact, memo string) (Transfers, error) {
	return Transfers{BaseOperation: NewBaseOperationFromFact(TransfersHint, fact, memo)}, nil
}

func (op *Transfers) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
