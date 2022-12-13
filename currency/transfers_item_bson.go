package currency // nolint:dupl

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseTransfersItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(bsonenc.NewHintedDoc(it.Hint()),
			bson.M{
				"receiver": it.receiver,
				"amounts":  it.amounts,
			}),
	)
}

type TransfersItemBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	RC string    `bson:"receiver"`
	AM bson.Raw  `bson:"amounts"`
}

func (it *BaseTransfersItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseTransfersItem")

	var uca TransfersItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uca); err != nil {
		return e(err, "")
	}

	it.BaseHinter = hint.NewBaseHinter(uca.HT)

	switch a, err := base.DecodeAddress(uca.RC, enc); {
	case err != nil:
		return e(err, "")
	default:
		it.receiver = a
	}

	ham, err := enc.DecodeSlice(uca.AM)
	if err != nil {
		return err
	}

	amounts := make([]Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(Amount)
		if !ok {
			return e(util.ErrWrongType.Errorf("expected Amount, not %T", ham[i]), "")
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
