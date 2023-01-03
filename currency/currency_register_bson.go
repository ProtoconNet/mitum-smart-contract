package currency // nolint: dupl

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fact CurrencyRegisterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"currency": fact.currency,
			},
			fact.BaseFact.BSONM(),
		))
}

type CurrencyRegisterFactBSONUnmarshaler struct {
	HT hint.Hint       `bson:"_hint"`
	CR json.RawMessage `bson:"currency"`
}

func (fact *CurrencyRegisterFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of CurrencyRegisterFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return e(err, "")
	}

	fact.BaseFact = ubf

	var uf CurrencyRegisterFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(uf.HT)

	return fact.unpack(enc, uf.CR)
}

func (op *CurrencyRegister) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo base.BaseNodeOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
