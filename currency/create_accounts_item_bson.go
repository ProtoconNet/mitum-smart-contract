package currency // nolint:dupl

import (
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateAccountsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   it.Hint().String(),
			"keys":    it.keys,
			"amounts": it.amounts,
		},
	)
}

type CreateAccountsItemBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	KS bson.Raw `bson:"keys"`
	AM bson.Raw `bson:"amounts"`
}

func (it *BaseCreateAccountsItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseCreateAccountsItem")

	var uit CreateAccountsItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uit.HT)
	if err != nil {
		return e(err, "")
	}

	return it.unpack(enc, ht, uit.KS, uit.AM)
}
