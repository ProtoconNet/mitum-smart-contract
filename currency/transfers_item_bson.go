package currency // nolint:dupl

import (
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseTransfersItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"receiver": it.receiver,
				"amounts":  it.amounts,
			}),
	)
}

type TransfersItemBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	RC string    `bson:"receiver"`
	AM bson.Raw  `bson:"amounts"`
}

func (it *BaseTransfersItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseTransfersItem")

	var uit TransfersItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.HT, uit.RC, uit.AM)
}
