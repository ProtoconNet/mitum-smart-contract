package digest

import (
	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (va AccountValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(va.Hint()),
		bson.M{
			"ac":      va.ac,
			"balance": va.balance,
			"height":  va.height,
		},
	))
}

type AccountValueBSONUnmarshaler struct {
	HT hint.Hint   `bson:"_hint"`
	AC bson.Raw    `bson:"ac"`
	BL bson.Raw    `bson:"balance"`
	H  base.Height `bson:"height"`
}

func (va *AccountValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uva AccountValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	return va.unpack(enc, uva.HT, uva.AC, uva.BL, uva.H)
}
