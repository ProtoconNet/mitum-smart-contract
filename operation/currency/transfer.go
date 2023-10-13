package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	TransferFactHint = hint.MustNewHint("mitum-currency-transfer-operation-fact-v0.0.1")
	TransferHint     = hint.MustNewHint("mitum-currency-transfer-operation-v0.0.1")
)

var MaxTransferItems uint = 10

type TransferItem interface {
	hint.Hinter
	util.IsValider
	AmountsItem
	Bytes() []byte
	Receiver() base.Address
	Rebuild() TransferItem
}

type TransferFact struct {
	base.BaseFact
	sender base.Address
	items  []TransferItem
}

func NewTransferFact(
	token []byte,
	sender base.Address,
	items []TransferItem,
) TransferFact {
	bf := base.NewBaseFact(TransferFactHint, token)
	fact := TransferFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact TransferFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact TransferFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact TransferFact) Bytes() []byte {
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

func (fact TransferFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
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
			return util.ErrInvalid.Errorf("duplicated receiver found, %v", it.Receiver())
		case fact.sender.Equal(it.Receiver()):
			return util.ErrInvalid.Errorf("receiver is same with sender, %v", fact.sender)
		default:
			foundReceivers[k] = struct{}{}
		}
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	return nil
}

func (fact TransferFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact TransferFact) Sender() base.Address {
	return fact.sender
}

func (fact TransferFact) Items() []TransferItem {
	return fact.items
}

func (fact TransferFact) Rebuild() TransferFact {
	items := make([]TransferItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.SetHash(fact.Hash())

	return fact
}

func (fact TransferFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)
	for i := range fact.items {
		as[i] = fact.items[i].Receiver()
	}

	as[len(fact.items)] = fact.Sender()

	return as, nil
}

type Transfer struct {
	common.BaseOperation
}

func NewTransfer(fact TransferFact) (Transfer, error) {
	return Transfer{BaseOperation: common.NewBaseOperation(TransferHint, fact)}, nil
}

func (op *Transfer) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
