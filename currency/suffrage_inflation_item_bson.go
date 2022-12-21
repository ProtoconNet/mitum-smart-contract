package currency // nolint:dupl

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"go.mongodb.org/mongo-driver/bson"
)

func (it SuffrageInflationItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bson.M{
				"receiver": it.receiver,
				"amount":   it.amount,
			}),
	)
}

type SuffrageInflationItemBSONUnmarshaler struct {
	RC string   `bson:"receiver"`
	AM bson.Raw `bson:"amount"`
}

func (it *SuffrageInflationItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageInflationItem")

	var uit SuffrageInflationItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	switch a, err := base.DecodeAddress(uit.RC, enc); {
	case err != nil:
		return err
	default:
		it.receiver = a
	}

	if hinter, err := enc.Decode(uit.AM); err != nil {
		return err
	} else if am, ok := hinter.(Amount); !ok {
		return e(util.ErrWrongType.Errorf("expected Amount not %T,", hinter), "")
	} else {
		it.amount = am
	}

	return nil
}
