package digest

import (
	"time"

	"github.com/spikeekips/mitum/base"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (va OperationValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(va.Hint()),
		bson.M{
			"op":           va.op,
			"height":       va.height,
			"confirmed_at": va.confirmedAt,
			"in_state":     va.inState,
			"reason":       va.reason,
			"index":        va.index,
		},
	))
}

type OperationValueBSONUnmarshaler struct {
	HT hint.Hint   `bson:"_hint"`
	OP bson.Raw    `bson:"op"`
	H  base.Height `bson:"height"`
	CT time.Time   `bson:"confirmed_at"`
	IN bool        `bson:"in_state"`
	RS bson.Raw    `bson:"reason"`
	ID uint64      `bson:"index"`
}

func (va *OperationValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uva OperationValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	va.BaseHinter = hint.NewBaseHinter(uva.HT)

	var op base.BaseOperation
	if err := op.DecodeBSON(uva.OP, enc); err != nil {
		return err
	}
	va.op = op

	var reason base.BaseOperationProcessReasonError

	if err := reason.DecodeBSON(uva.RS, enc); err != nil {
		return err
	}

	va.height = uva.H
	va.confirmedAt = uva.CT
	va.inState = uva.IN
	va.index = uva.ID
	va.reason = reason
	return nil
}
