package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ky BaseAccountKey) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  ky.Hint().String(),
			"weight": ky.w,
			"key":    ky.k.String(),
		},
	)
}

type KeyBSONUnmarshaler struct {
	HT string `bson:"_hint"`
	W  uint   `bson:"weight"`
	K  string `bson:"key"`
}

func (ky *BaseAccountKey) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseAccountKey")

	var uk KeyBSONUnmarshaler
	if err := bson.Unmarshal(b, &uk); err != nil {
		return e(err, "")
	}
	ht, err := hint.ParseHint(uk.HT)
	if err != nil {
		return e(err, "")
	}

	return ky.unpack(enc, ht, uk.W, uk.K)
}

func (ks BaseAccountKeys) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     ks.Hint().String(),
			"hash":      ks.Hash().String(),
			"keys":      ks.keys,
			"threshold": ks.threshold,
		},
	)
}

type KeysBSONUnmarshaler struct {
	HT string   `bson:"_hint"`
	HS string   `bson:"hash"`
	KS bson.Raw `bson:"keys"`
	TH uint     `bson:"threshold"`
}

func (ks *BaseAccountKeys) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseAccountKeys")

	var uks KeysBSONUnmarshaler
	if err := bson.Unmarshal(b, &uks); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uks.HT)
	if err != nil {
		return e(err, "")
	}

	var vh valuehash.HashDecoder
	err = vh.UnmarshalText(valuehash.NewBytesFromString(uks.HS))
	if err != nil {
		return e(err, "")
	}

	return ks.unpack(enc, ht, vh, uks.KS, uks.TH)
}
