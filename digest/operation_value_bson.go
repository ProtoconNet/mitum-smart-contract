package digest

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"time"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
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
	Hint        string      `bson:"_hint"`
	OP          bson.Raw    `bson:"op"`
	Height      base.Height `bson:"height"`
	ConfirmedAt time.Time   `bson:"confirmed_at"`
	InState     bool        `bson:"in_state"`
	RS          string      `bson:"reason"`
	Index       uint64      `bson:"index"`
}

func (va *OperationValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of OperationValue")
	var uva OperationValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uva.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	va.BaseHinter = hint.NewBaseHinter(ht)

	var op common.BaseOperation
	if err := op.DecodeBSON(uva.OP, enc); err != nil {
		return e.Wrap(err)
	}

	va.op = op

	va.height = uva.Height
	va.confirmedAt = uva.ConfirmedAt
	va.inState = uva.InState
	va.index = uva.Index
	va.reason = uva.RS
	return nil
}
