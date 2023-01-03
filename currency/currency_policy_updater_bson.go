package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fact CurrencyPolicyUpdaterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"currency": fact.currency,
				"policy":   fact.policy,
			},
			fact.BaseFact.BSONM(),
		))
}

type CurrencyPolicyUpdaterFactBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	CR string    `bson:"currency"`
	PO bson.Raw  `bson:"policy"`
}

func (fact *CurrencyPolicyUpdaterFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of CurrencyPolicyUpdaterFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return e(err, "")
	}

	fact.BaseFact = ubf

	var uf CurrencyPolicyUpdaterFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(uf.HT)

	return fact.unpack(enc, uf.CR, uf.PO)
}

func (op *CurrencyPolicyUpdater) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo base.BaseNodeOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
