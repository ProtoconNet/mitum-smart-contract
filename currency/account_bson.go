package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
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
	var uac AccountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return err
	}

	return ac.unpack(enc, uac.H, uac.AD, uac.KS)
}
