package currency // nolint:dupl

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseTransferItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    it.Hint().String(),
			"receiver": it.receiver,
			"amounts":  it.amounts,
		},
	)
}

type TransferItemBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Receiver string   `bson:"receiver"`
	Amounts  bson.Raw `bson:"amounts"`
}

func (it *BaseTransferItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of BaseTransferItem")

	var uit TransferItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, ht, uit.Receiver, uit.Amounts)
}
