package currency // nolint: dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact SuffrageInflationFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": fact.Hint().String(),
			"items": fact.items,
			"hash":  fact.BaseFact.Hash().String(),
			"token": fact.BaseFact.Token(),
		},
	)
}

type SuffrageInflationFactBSONUnmarshaler struct {
	Hint  string     `bson:"_hint"`
	Items []bson.Raw `bson:"items"`
}

func (fact *SuffrageInflationFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageInflationFact")

	var u base.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf SuffrageInflationFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e(err, "")
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	items := make([]SuffrageInflationItem, len(uf.Items))
	for i := range uf.Items {
		item := SuffrageInflationItem{}
		if err := item.DecodeBSON(uf.Items[i], enc); err != nil {
			return e(err, "")
		}
		items[i] = item
	}

	fact.items = items

	return nil
}

func (op *SuffrageInflation) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageInflation")

	var ubo base.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo

	return nil
}
