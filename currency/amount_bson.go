package currency

import (
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (am Amount) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(am.Hint()),
		bson.M{
			"currency": am.cid,
			"amount":   am.big,
		}),
	)
}

type AmountBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	CR string    `bson:"currency"`
	BG Big       `bson:"amount"`
}

func (am *Amount) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal bson of Amount")

	var uam AmountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uam); err != nil {
		return e(err, "")
	}

	am.BaseHinter = hint.NewBaseHinter(uam.HT)
	am.big = uam.BG
	am.cid = CurrencyID(uam.CR)

	return nil
}
