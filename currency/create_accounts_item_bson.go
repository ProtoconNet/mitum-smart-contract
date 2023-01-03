package currency // nolint:dupl

import (
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateAccountsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"keys":    it.keys,
				"amounts": it.amounts,
			}),
	)
}

type CreateAccountsItemBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	KS bson.Raw  `bson:"keys"`
	AM bson.Raw  `bson:"amounts"`
}

func (it *BaseCreateAccountsItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseCreateAccountsItem")

	var uit CreateAccountsItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.HT, uit.KS, uit.AM)
}
