package currency // nolint:dupl

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v2/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it SuffrageInflationItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    it.Hint().String(),
			"receiver": it.receiver,
			"amount":   it.amount,
		},
	)
}

type SuffrageInflationItemBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Receiver string   `bson:"receiver"`
	Amount   bson.Raw `bson:"amount"`
}

func (it *SuffrageInflationItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageInflationItem")

	var uit SuffrageInflationItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}
	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return e(err, "")
	}

	return it.unpack(enc, ht, uit.Receiver, uit.Amount)
}
