package digest

import (
	"time"

	"github.com/spikeekips/mitum-currency/currency"
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (va OperationValue) MarshalBSON() ([]byte, error) {
	op := map[string]interface{}{
		"_hint": va.op.Hint().String(),
		"hash":  va.op.Hash().String(),
		"fact":  va.op.Fact(),
		"signs": va.op.Signs(),
	}
	return bsonenc.Marshal(
		bson.M{
			"_hint":        va.Hint().String(),
			"op":           op,
			"height":       va.height,
			"confirmed_at": va.confirmedAt,
			"in_state":     va.inState,
			"reason":       va.reason,
			"index":        va.index,
		},
	)
}

type OperationValueBSONUnmarshaler struct {
	HT string      `bson:"_hint"`
	OP bson.Raw    `bson:"op"`
	H  base.Height `bson:"height"`
	CT time.Time   `bson:"confirmed_at"`
	IN bool        `bson:"in_state"`
	//RS bson.Raw    `bson:"reason"`
	ID uint64 `bson:"index"`
}

func (va *OperationValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of OperationValue")
	var uva OperationValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uva.HT)
	if err != nil {
		return e(err, "")
	}

	va.BaseHinter = hint.NewBaseHinter(ht)

	var op currency.BaseOperation
	if err := op.DecodeBSON(uva.OP, enc); err != nil {
		return e(err, "")
	}

	va.op = op

	// var reason base.BaseOperationProcessReasonError

	// if err := reason.DecodeBSON(uva.RS, enc); err != nil {
	// 	return err
	// }

	va.height = uva.H
	va.confirmedAt = uva.CT
	va.inState = uva.IN
	va.index = uva.ID
	// va.reason = reason
	return nil
}
