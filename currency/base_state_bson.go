package currency

import (
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (s BaseState) BSONM() bson.M {
	return bson.M{
		"_hint":      s.Hint().String(),
		"hash":       s.h,
		"previous":   s.previous,
		"value":      s.v,
		"key":        s.k,
		"operations": s.ops,
		"height":     s.height,
	}
}

func (s BaseState) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		s.BSONM(),
	)
}

type BaseStateBSONUnmarshaler struct {
	Hint       string            `bson:"_hint"`
	Hash       valuehash.Bytes   `bson:"hash"`
	Previous   valuehash.Bytes   `bson:"previous"`
	Key        string            `bson:"key"`
	Value      bson.Raw          `bson:"value"`
	Operations []valuehash.Bytes `bson:"operations"`
	Height     base.Height       `bson:"height"`
}

func (s *BaseState) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal BaseState")

	var u BaseStateBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e(err, "")
	}
	s.BaseHinter = hint.NewBaseHinter(ht)

	s.h = u.Hash
	s.previous = u.Previous
	s.height = u.Height
	s.k = u.Key

	s.ops = make([]util.Hash, len(u.Operations))

	for i := range u.Operations {
		s.ops[i] = u.Operations[i]
	}

	switch i, err := DecodeStateValue(u.Value, enc); {
	case err != nil:
		return e(err, "")
	default:
		s.v = i
	}

	return nil
}
