package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	CreateContractAccountFactHint = hint.MustNewHint("mitum-currency-create-contract-account-operation-fact-v0.0.1")
	CreateContractAccountHint     = hint.MustNewHint("mitum-currency-create-contract-account-operation-v0.0.1")
)

var MaxCreateContractAccountItems uint = 10

type CreateContractAccountItem interface {
	hint.Hinter
	util.IsValider
	currency.AmountsItem
	Bytes() []byte
	Keys() types.AccountKeys
	Address() (base.Address, error)
	Rebuild() CreateContractAccountItem
	AddressType() hint.Type
}

type CreateContractAccountFact struct {
	base.BaseFact
	sender base.Address
	items  []CreateContractAccountItem
}

func NewCreateContractAccountFact(token []byte, sender base.Address, items []CreateContractAccountItem) CreateContractAccountFact {
	bf := base.NewBaseFact(CreateContractAccountFactHint, token)
	fact := CreateContractAccountFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CreateContractAccountFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CreateContractAccountFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateContractAccountFact) Bytes() []byte {
	is := make([][]byte, len(fact.items))
	for i := range fact.items {
		is[i] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(
		fact.Token(),
		fact.sender.Bytes(),
		util.ConcatBytesSlice(is...),
	)
}

func (fact CreateContractAccountFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.items); n < 1 {
		return util.ErrInvalid.Errorf("empty items")
	} else if n > int(MaxCreateContractAccountItems) {
		return util.ErrInvalid.Errorf("items, %d over max, %d", n, MaxCreateContractAccountItems)
	}

	if err := util.CheckIsValiders(nil, false, fact.sender); err != nil {
		return err
	}

	foundKeys := map[string]struct{}{}
	for i := range fact.items {
		if err := util.CheckIsValiders(nil, false, fact.items[i]); err != nil {
			return err
		}

		it := fact.items[i]
		k := it.Keys().Hash().String()
		if _, found := foundKeys[k]; found {
			return util.ErrInvalid.Errorf("duplicate acocunt Keys found, %s", k)
		}

		switch a, err := it.Address(); {
		case err != nil:
			return err
		case fact.sender.Equal(a):
			return util.ErrInvalid.Errorf("target address is same with sender, %v", fact.sender)
		default:
			foundKeys[k] = struct{}{}
		}
	}

	return nil
}

func (fact CreateContractAccountFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact CreateContractAccountFact) Sender() base.Address {
	return fact.sender
}

func (fact CreateContractAccountFact) Items() []CreateContractAccountItem {
	return fact.items
}

func (fact CreateContractAccountFact) Targets() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items))
	for i := range fact.items {
		a, err := fact.items[i].Address()
		if err != nil {
			return nil, err
		}
		as[i] = a
	}

	return as, nil
}

func (fact CreateContractAccountFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)

	tas, err := fact.Targets()
	if err != nil {
		return nil, err
	}
	copy(as, tas)

	as[len(fact.items)] = fact.sender

	return as, nil
}

func (fact CreateContractAccountFact) Rebuild() CreateContractAccountFact {
	items := make([]CreateContractAccountItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.SetHash(fact.GenerateHash())

	return fact
}

type CreateContractAccount struct {
	common.BaseOperation
}

func NewCreateContractAccount(fact CreateContractAccountFact) (CreateContractAccount, error) {
	return CreateContractAccount{BaseOperation: common.NewBaseOperation(CreateContractAccountHint, fact)}, nil
}

func (op *CreateContractAccount) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
