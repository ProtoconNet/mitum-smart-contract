package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ac Account) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(ac.Hint()),
		bson.M{
			"hash":    ac.h,
			"address": ac.address,
			"keys":    ac.keys,
		},
	))
}

type AccountBSONUnmarshaler struct {
	HT hint.Hint             `bson:"_hint"`
	H  valuehash.HashDecoder `bson:"hash"`
	AD string                `bson:"address"`
	KS bson.Raw              `bson:"keys"`
}

func (ac *Account) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of Account")

	var uac AccountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return e(err, "")
	}

	ac.BaseHinter = hint.NewBaseHinter(uac.HT)

	return ac.unpack(enc, uac.H, uac.AD, uac.KS)
}
