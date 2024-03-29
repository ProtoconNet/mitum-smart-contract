package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

var (
	MintFactHint = hint.MustNewHint("mitum-currency-mint-operation-fact-v0.0.1")
	MintHint     = hint.MustNewHint("mitum-currency-mint-operation-v0.0.1")
)

var maxMintItem = 10

type MintFact struct {
	base.BaseFact
	items []MintItem
}

func NewMintFact(token []byte, items []MintItem) MintFact {
	fact := MintFact{
		BaseFact: base.NewBaseFact(MintFactHint, token),
		items:    items,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact MintFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact MintFact) Bytes() []byte {
	bi := make([][]byte, len(fact.items)+1)
	bi[0] = fact.Token()

	for i := range fact.items {
		bi[i+1] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(bi...)
}

func (fact MintFact) IsValid(b []byte) error {
	if err := fact.BaseHinter.IsValid(nil); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	switch n := len(fact.items); {
	case n < 1:
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("Empty items for MintFact")))
	case n > maxMintItem:
		return common.ErrFactInvalid.Wrap(common.ErrArrayLen.Wrap(errors.Errorf("Array length: Too many items; %d > %d", n, maxMintItem)))
	}

	founds := map[string]struct{}{}
	for i := range fact.items {
		item := fact.items[i]
		if err := item.IsValid(nil); err != nil {
			return common.ErrFactInvalid.Wrap(err)
		}

		k := item.receiver.String() + "-" + item.amount.Currency().String()
		if _, found := founds[k]; found {
			return common.ErrFactInvalid.Wrap(common.ErrDupVal.Wrap(errors.Errorf("Duplicated value: Item in MintFact")))
		}
		founds[k] = struct{}{}
	}

	if err := common.IsValidOperationFact(fact, b); err != nil {
		return common.ErrFactInvalid.Wrap(err)
	}

	return nil
}

func (fact MintFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact MintFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact MintFact) Items() []MintItem {
	return fact.items
}

type Mint struct {
	common.BaseNodeOperation
}

func NewMint(
	fact MintFact,
) (Mint, error) {
	return Mint{BaseNodeOperation: common.NewBaseNodeOperation(MintHint, fact)}, nil
}
