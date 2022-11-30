package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ky BaseAccountKey) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(ky.Hint()),
		bson.M{
			"weight": ky.w,
			"key":    ky.k.String(),
		},
	))
}

type KeyBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	W  uint      `bson:"weight"`
	K  string    `bson:"key"`
}

func (ky *BaseAccountKey) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal bson of BaseAccountKey")

	var uk KeyBSONUnmarshaler
	if err := bson.Unmarshal(b, &uk); err != nil {
		return e(err, "")
	}

	return ky.unpack(enc, uk.HT, uk.W, uk.K)
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
	HT hint.Hint       `bson:"_hint"`
	H  valuehash.Bytes `bson:"hash"`
	KS bson.Raw        `bson:"keys"`
	TH uint            `bson:"threshold"`
}

func (ks *BaseAccountKeys) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal bson of BaseAccountKeys")

	var uks KeysBSONUnmarshaler
	if err := bson.Unmarshal(b, &uks); err != nil {
		return e(err, "")
	}

	ks.BaseHinter = hint.NewBaseHinter(uks.HT)
	hks, err := enc.DecodeSlice(uks.KS)
	if err != nil {
		return err
	}

	keys := make([]AccountKey, len(hks))
	for i := range hks {
		j, ok := hks[i].(BaseAccountKey)
		if !ok {
			return util.ErrWrongType.Errorf("expected Key, not %T", hks[i])
		}

		keys[i] = j
	}
	ks.keys = keys

	ks.h = uks.H
	ks.threshold = uks.TH

	return nil
}
