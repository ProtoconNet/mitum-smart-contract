package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	CreateContractAccountsFactHint = hint.MustNewHint("mitum-currency-create-contract-accounts-operation-fact-v0.0.1")
	CreateContractAccountsHint     = hint.MustNewHint("mitum-currency-create-contract-accounts-operation-v0.0.1")
)

var MaxCreateContractAccountsItems uint = 10

type CreateContractAccountsItem interface {
	hint.Hinter
	util.IsValider
	currency.AmountsItem
	Bytes() []byte
	Keys() base.AccountKeys
	Address() (mitumbase.Address, error)
	Rebuild() CreateContractAccountsItem
	AddressType() hint.Type
}

type CreateContractAccountsFact struct {
	mitumbase.BaseFact
	sender mitumbase.Address
	items  []CreateContractAccountsItem
}

func NewCreateContractAccountsFact(token []byte, sender mitumbase.Address, items []CreateContractAccountsItem) CreateContractAccountsFact {
	bf := mitumbase.NewBaseFact(CreateContractAccountsFactHint, token)
	fact := CreateContractAccountsFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CreateContractAccountsFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CreateContractAccountsFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateContractAccountsFact) Bytes() []byte {
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

func (fact CreateContractAccountsFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := base.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.items); n < 1 {
		return util.ErrInvalid.Errorf("empty items")
	} else if n > int(MaxCreateContractAccountsItems) {
		return util.ErrInvalid.Errorf("items, %d over max, %d", n, MaxCreateContractAccountsItems)
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
			return util.ErrInvalid.Errorf("target address is same with sender, %q", fact.sender)
		default:
			foundKeys[k] = struct{}{}
		}
	}

	return nil
}

func (fact CreateContractAccountsFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact CreateContractAccountsFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact CreateContractAccountsFact) Items() []CreateContractAccountsItem {
	return fact.items
}

func (fact CreateContractAccountsFact) Targets() ([]mitumbase.Address, error) {
	as := make([]mitumbase.Address, len(fact.items))
	for i := range fact.items {
		a, err := fact.items[i].Address()
		if err != nil {
			return nil, err
		}
		as[i] = a
	}

	return as, nil
}

func (fact CreateContractAccountsFact) Addresses() ([]mitumbase.Address, error) {
	as := make([]mitumbase.Address, len(fact.items)+1)

	tas, err := fact.Targets()
	if err != nil {
		return nil, err
	}
	copy(as, tas)

	as[len(fact.items)] = fact.sender

	return as, nil
}

func (fact CreateContractAccountsFact) Rebuild() CreateContractAccountsFact {
	items := make([]CreateContractAccountsItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.SetHash(fact.GenerateHash())

	return fact
}

type CreateContractAccounts struct {
	base.BaseOperation
}

func NewCreateContractAccounts(fact CreateContractAccountsFact) (CreateContractAccounts, error) {
	return CreateContractAccounts{BaseOperation: base.NewBaseOperation(CreateContractAccountsHint, fact)}, nil
}

func (op *CreateContractAccounts) HashSign(priv mitumbase.Privatekey, networkID mitumbase.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
