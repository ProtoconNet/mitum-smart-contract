package currency

import (
	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
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
	CR string `bson:"currency"`
	BG Big    `bson:"amount"`
}

func (am *Amount) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uam AmountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uam); err != nil {
		return err
	}

	am.big = uam.BG
	am.cid = CurrencyID(uam.CR)

	return nil
}
