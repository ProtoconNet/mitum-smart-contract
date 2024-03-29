package types

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (am Amount) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    am.Hint().String(),
			"currency": am.cid,
			"amount":   am.big.String(),
		},
	)
}

type AmountBSONUnmarshaler struct {
	Hint      string `bson:"_hint"`
	Currency  string `bson:"currency"`
	AmountBig string `bson:"amount"`
}

func (am *Amount) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of Amount")

	var uam AmountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uam); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uam.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	am.BaseHinter = hint.NewBaseHinter(ht)

	return am.unpack(enc, uam.Currency, uam.AmountBig)
}
