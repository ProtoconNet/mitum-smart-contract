package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fact TransfersFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"sender": fact.sender,
				"items":  fact.items,
			},
			fact.BaseFact.BSONM(),
		))
}

type TransfersFactBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	SD string    `bson:"sender"`
	IT bson.Raw  `bson:"items"`
}

func (fact *TransfersFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of TransfersFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return err
	}

	fact.BaseFact = ubf

	var utf TransfersFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &utf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(utf.HT)
	switch a, err := base.DecodeAddress(utf.SD, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(utf.IT)
	if err != nil {
		return e(err, "")
	}

	items := make([]TransfersItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(TransfersItem)
		if !ok {
			return util.ErrWrongType.Errorf("expected TransfersItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}

func (op *Transfers) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
