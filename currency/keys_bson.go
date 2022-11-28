package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ky BaseAccountKey) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(ky.Hint()),
		bson.M{
			"weight": ky.w,
			"key":    ky.k,
		},
	))
}

type KeyBSONUnmarshaler struct {
	W uint   `bson:"weight"`
	K string `bson:"key"`
}

func (ky *BaseAccountKey) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal bson of BaseAccountKey")

	var uk KeyBSONUnmarshaler
	if err := bson.Unmarshal(b, &uk); err != nil {
		return e(err, "")
	}

	return ky.unpack(enc, uk.W, uk.K)
}

func (ks BaseAccountKeys) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(ks.Hint()),
		bson.M{
			"hash":      ks.h,
			"keys":      ks.keys,
			"threshold": ks.threshold,
		},
	))
}

type KeysBSONUnmarshaler struct {
	H  valuehash.HashDecoder `bson:"hash"`
	KS bson.Raw              `bson:"keys"`
	TH uint                  `bson:"threshold"`
}

func (ks *BaseAccountKeys) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal bson of BaseAccountKeys")

	var uks KeysBSONUnmarshaler
	if err := bson.Unmarshal(b, &uks); err != nil {
		return e(err, "")
	}

	return ks.unpack(enc, uks.H, uks.KS, uks.TH)
}
