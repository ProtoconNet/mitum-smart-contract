package digest

import (
	"time"

	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
	"github.com/spikeekips/mitum/base"
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
	OP bson.Raw    `bson:"op"`
	HT base.Height `bson:"height"`
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

	if err := enc.Unmarshal(uva.OP, &va.op); err != nil {
		return err
	}

	if err := enc.Unmarshal(uva.RS, &va.reason); err != nil {
		return err
	}

	va.height = uva.HT
	va.confirmedAt = uva.CT
	va.inState = uva.IN
	va.index = uva.ID

	return nil
}
