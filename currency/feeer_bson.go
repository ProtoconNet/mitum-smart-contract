package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
)

func (fa NilFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.NewHintedDoc(fa.Hint()))
}

func (fa *NilFeeer) UnmarsahlBSON(b []byte) error {
	e := util.StringErrorFunc("failed to unmarshal bson of NilFeeer")

	var head bsonenc.HintedHead
	if err := bsonenc.Unmarshal(b, &head); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(head.H)
	if err != nil {
		return e(err, "")
	}

	fa.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}

func (fa FixedFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    fa.Hint().String(),
			"receiver": fa.receiver,
			"amount":   fa.amount.String(),
		},
	)

}

type FixedFeeerBSONUnpacker struct {
	HT string `bson:"_hint"`
	RC string `bson:"receiver"`
	AM string `bson:"amount"`
}

func (fa *FixedFeeer) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of FixedFeeer")

	var ufa FixedFeeerBSONUnpacker
	if err := enc.Unmarshal(b, &ufa); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(ufa.HT)
	if err != nil {
		return e(err, "")
	}

	return fa.unpack(enc, ht, ufa.RC, ufa.AM)
}

func (fa RatioFeeer) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    fa.Hint().String(),
			"receiver": fa.receiver,
			"ratio":    fa.ratio,
			"min":      fa.min.String(),
			"max":      fa.max.String(),
		},
	)
}

type RatioFeeerBSONUnpacker struct {
	HT string  `bson:"_hint"`
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
	ht, err := hint.ParseHint(ufa.HT)
	if err != nil {
		return e(err, "")
	}

	return fa.unpack(enc, ht, ufa.RC, ufa.RA, ufa.MI, ufa.MA)
}
