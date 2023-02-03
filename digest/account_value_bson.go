package digest

import (
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (va AccountValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bson.M{
			"_hint":   va.Hint().String(),
			"ac":      va.ac,
			"balance": va.balance,
			"height":  va.height,
		},
	))
}

type AccountValueBSONUnmarshaler struct {
	HT string      `bson:"_hint"`
	AC bson.Raw    `bson:"ac"`
	BL bson.Raw    `bson:"balance"`
	H  base.Height `bson:"height"`
}

func (va *AccountValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of AccountValue")

	var uva AccountValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uva.HT)
	if err != nil {
		return e(err, "")
	}

	return va.unpack(enc, ht, uva.AC, uva.BL, uva.H)
}
