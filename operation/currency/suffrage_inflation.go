package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	SuffrageInflationFactHint = hint.MustNewHint("mitum-currency-suffrage-inflation-operation-fact-v0.0.1")
	SuffrageInflationHint     = hint.MustNewHint("mitum-currency-suffrage-inflation-operation-v0.0.1")
)

var maxSuffrageInflationItem = 10

type SuffrageInflationFact struct {
	base.BaseFact
	items []SuffrageInflationItem
}

func NewSuffrageInflationFact(token []byte, items []SuffrageInflationItem) SuffrageInflationFact {
	fact := SuffrageInflationFact{
		BaseFact: base.NewBaseFact(SuffrageInflationFactHint, token),
		items:    items,
	}

	fact.SetHash(fact.GenerateHash())

	return fact
}

func (fact SuffrageInflationFact) Hash() util.Hash {
	return fact.BaseFact.Hash()
}

func (fact SuffrageInflationFact) Bytes() []byte {
	bi := make([][]byte, len(fact.items)+1)
	bi[0] = fact.Token()

	for i := range fact.items {
		bi[i+1] = fact.items[i].Bytes()
	}

	return util.ConcatBytesSlice(bi...)
}

func (fact SuffrageInflationFact) IsValid(b []byte) error {
	if err := common.IsValidOperationFact(fact, b); err != nil {
		return err
	}

	switch l := len(fact.items); {
	case l < 1:
		return util.ErrInvalid.Errorf("empty items for SuffrageInflationFact")
	case l > maxSuffrageInflationItem:
		return util.ErrInvalid.Errorf("too many items; %d > %d", l, maxSuffrageInflationItem)
	}

	founds := map[string]struct{}{}
	for i := range fact.items {
		item := fact.items[i]
		if err := item.IsValid(nil); err != nil {
			return util.ErrInvalid.Errorf("invalid SuffrageInflationItem: %v", err)
		}

		k := item.receiver.String() + "-" + item.amount.Currency().String()
		if _, found := founds[k]; found {
			return util.ErrInvalid.Errorf("duplicated item found in SuffrageInflationFact")
		}
		founds[k] = struct{}{}
	}

	return nil
}

func (fact SuffrageInflationFact) GenerateHash() util.Hash {
	return valuehash.NewSHA256(fact.Bytes())
}

func (fact SuffrageInflationFact) Token() base.Token {
	return fact.BaseFact.Token()
}

func (fact SuffrageInflationFact) Items() []SuffrageInflationItem {
	return fact.items
}

type SuffrageInflation struct {
	common.BaseNodeOperation
}

func NewSuffrageInflation(
	fact SuffrageInflationFact,
) (SuffrageInflation, error) {
	return SuffrageInflation{BaseNodeOperation: common.NewBaseNodeOperation(SuffrageInflationHint, fact)}, nil
}
