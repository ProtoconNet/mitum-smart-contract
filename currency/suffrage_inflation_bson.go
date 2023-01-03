package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fact SuffrageInflationFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"items": fact.items,
			},
			fact.BaseFact.BSONM(),
		))
}

type SuffrageInflationFactBSONUnmarshaler struct {
	HT hint.Hint  `bson:"_hint"`
	IT []bson.Raw `bson:"items"`
}

func (fact *SuffrageInflationFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageInflationFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return e(err, "")
	}

	fact.BaseFact = ubf

	var uf SuffrageInflationFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(uf.HT)

	items := make([]SuffrageInflationItem, len(uf.IT))
	for i := range uf.IT {
		item := SuffrageInflationItem{}
		if err := item.DecodeBSON(uf.IT[i], enc); err != nil {
			return e(err, "")
		}
		items[i] = item
	}

	fact.items = items

	return nil
}

func (op *SuffrageInflation) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo base.BaseNodeOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
