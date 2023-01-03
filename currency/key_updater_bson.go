package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fact KeyUpdaterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"target":   fact.target,
				"keys":     fact.keys,
				"currency": fact.currency,
			},
			fact.BaseFact.BSONM(),
		))
}

type KeyUpdaterFactBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	TG string    `bson:"target"`
	KS bson.Raw  `bson:"keys"`
	CR string    `bson:"currency"`
}

func (fact *KeyUpdaterFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of KeyUpdaterFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return e(err, "")
	}

	fact.BaseFact = ubf

	var uf KeyUpdaterFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(uf.HT)

	return fact.unpack(enc, uf.TG, uf.KS, uf.CR)
}

func (op *KeyUpdater) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
