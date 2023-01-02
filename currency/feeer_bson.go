package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fa NilFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.NewHintedDoc(fa.Hint()))
}

func (fa *NilFeeer) UnmarsahlBSON(b []byte) error {
	e := util.StringErrorFunc("failed to unmarshal bson of NilFeeer")

	var ht bsonenc.HintedHead
	if err := bsonenc.Unmarshal(b, &ht); err != nil {
		return e(err, "")
	}

	fa.BaseHinter = hint.NewBaseHinter(ht.H)

	return nil
}

func (fa FixedFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(fa.Hint()),
		bson.M{
			"receiver": fa.receiver,
			"amount":   fa.amount.String(),
		}),
	)
}

type FixedFeeerBSONUnpacker struct {
	RC string `bson:"receiver"`
	AM string `bson:"amount"`
}

func (fa *FixedFeeer) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of FixedFeeer")

	var ufa FixedFeeerBSONUnpacker
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e(err, "")
	}

	return fa.unpack(enc, ufa.RC, ufa.AM)
}

func (fa RatioFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(fa.Hint()),
		bson.M{
			"receiver": fa.receiver,
			"ratio":    fa.ratio,
			"min":      fa.min.String(),
			"max":      fa.max.String(),
		}),
	)
}

type RatioFeeerBSONUnpacker struct {
	RC string  `bson:"receiver"`
	RA float64 `bson:"ratio"`
	MI string  `bson:"min"`
	MA string  `bson:"max"`
}

func (fa *RatioFeeer) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of RatioFeeer")

	var ufa RatioFeeerBSONUnpacker
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e(err, "")
	}

	return fa.unpack(enc, ufa.RC, ufa.RA, ufa.MI, ufa.MA)
}
