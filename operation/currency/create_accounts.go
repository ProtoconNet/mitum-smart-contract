package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	CreateAccountsFactHint = hint.MustNewHint("mitum-currency-create-accounts-operation-fact-v0.0.1")
	CreateAccountsHint     = hint.MustNewHint("mitum-currency-create-accounts-operation-v0.0.1")
)

var MaxCreateAccountsItems uint = 1000

type AmountsItem interface {
	Amounts() []base.Amount
}

type CreateAccountsItem interface {
	hint.Hinter
	util.IsValider
	AmountsItem
	Bytes() []byte
	Keys() base.AccountKeys
	Address() (mitumbase.Address, error)
	Rebuild() CreateAccountsItem
	AddressType() hint.Type
}

type CreateAccountsFact struct {
	mitumbase.BaseFact
	sender mitumbase.Address
	items  []CreateAccountsItem
}

func NewCreateAccountsFact(
	token []byte,
	sender mitumbase.Address,
	items []CreateAccountsItem,
) CreateAccountsFact {
	bf := mitumbase.NewBaseFact(CreateAccountsFactHint, token)
	fact := CreateAccountsFact{
		BaseFact: bf,
		sender:   sender,
		items:    items,
	}
	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact CreateAccountsFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact CreateAccountsFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact CreateAccountsFact) Bytes() []byte {
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

func (fact CreateAccountsFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := base.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	if n := len(fact.items); n < 1 {
		return util.ErrInvalid.Errorf("empty items")
	} else if n > int(MaxCreateAccountsItems) {
		return util.ErrInvalid.Errorf("items, %d over max, %d", n, MaxCreateAccountsItems)
	}

	if err := util.CheckIsValiders(nil, false, fact.sender); err != nil {
		return err
	}

	foundKeys := map[string]struct{}{}
	for i := range fact.items {
		it := fact.items[i]
		if err := util.CheckIsValiders(nil, false, it); err != nil {
			return err
		}

		k := it.Keys().Hash().String()
		if _, found := foundKeys[k]; found {
			return util.ErrInvalid.Errorf("duplicated acocunt Keys found, %s", k)
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

func (fact CreateAccountsFact) Token() mitumbase.Token {
	return fact.BaseFact.Token()
}

func (fact CreateAccountsFact) Sender() mitumbase.Address {
	return fact.sender
}

func (fact CreateAccountsFact) Items() []CreateAccountsItem {
	return fact.items
}

func (fact CreateAccountsFact) Targets() ([]mitumbase.Address, error) {
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

func (fact CreateAccountsFact) Addresses() ([]mitumbase.Address, error) {
	as := make([]mitumbase.Address, len(fact.items)+1)

	tas, err := fact.Targets()
	if err != nil {
		return nil, err
	}
	copy(as, tas)

	as[len(fact.items)] = fact.Sender()

	return as, nil
}

func (fact CreateAccountsFact) Rebuild() CreateAccountsFact {
	items := make([]CreateAccountsItem, len(fact.items))
	for i := range fact.items {
		it := fact.items[i]
		items[i] = it.Rebuild()
	}

	fact.items = items
	fact.SetHash(fact.Hash())

	return fact
}

type CreateAccounts struct {
	base.BaseOperation
}

func NewCreateAccounts(fact CreateAccountsFact) (CreateAccounts, error) {
	return CreateAccounts{BaseOperation: base.NewBaseOperation(CreateAccountsHint, fact)}, nil
}

func (op *CreateAccounts) HashSign(priv mitumbase.Privatekey, networkID mitumbase.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
