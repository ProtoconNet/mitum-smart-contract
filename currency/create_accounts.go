package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

var (
	CreateAccountsFactHint = hint.MustNewHint("mitum-currency-create-accounts-operation-fact-v0.0.1")
	CreateAccountsHint     = hint.MustNewHint("mitum-currency-create-accounts-operation-v0.0.1")
)

var MaxCreateAccountsItems uint = 1000

type AmountsItem interface {
	Amounts() []Amount
}

type CreateAccountsItem interface {
	hint.Hinter
	util.IsValider
	AmountsItem
	Bytes() []byte
	Keys() AccountKeys
	Address() (base.Address, error)
	Rebuild() CreateAccountsItem
}

type CreateAccountsFact struct {
	base.BaseFact
	sender base.Address
	items  []CreateAccountsItem
}

func NewCreateAccountsFact(
	token []byte,
	sender base.Address,
	items []CreateAccountsItem,
) CreateAccountsFact {
	bf := base.NewBaseFact(CreateAccountsFactHint, token)
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

	if err := IsValidOperationFact(fact, b); err != nil {
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

func (fact CreateAccountsFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact CreateAccountsFact) Sender() base.Address {
	return fact.sender
}

func (fact CreateAccountsFact) Items() []CreateAccountsItem {
	return fact.items
}

func (fact CreateAccountsFact) Targets() ([]base.Address, error) {
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

func (fact CreateAccountsFact) Addresses() ([]base.Address, error) {
	as := make([]base.Address, len(fact.items)+1)

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
	BaseOperation
}

func NewCreateAccounts(fact CreateAccountsFact, memo string) (CreateAccounts, error) {
	return CreateAccounts{BaseOperation: NewBaseOperationFromFact(CreateAccountsHint, fact, memo)}, nil
}

func (op *CreateAccounts) HashSign(priv base.Privatekey, networkID base.NetworkID) error {
	err := op.Sign(priv, networkID)
	if err != nil {
		return err
	}
	return nil
}
