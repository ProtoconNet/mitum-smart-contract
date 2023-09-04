package currency // nolint:dupl

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it MintItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    it.Hint().String(),
			"receiver": it.receiver,
			"amount":   it.amount,
		},
	)
}

type MintItemBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Receiver string   `bson:"receiver"`
	Amount   bson.Raw `bson:"amount"`
}

func (it *MintItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("failed to decode bson of MintItem")

	var uit MintItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}
	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, ht, uit.Receiver, uit.Amount)
}
