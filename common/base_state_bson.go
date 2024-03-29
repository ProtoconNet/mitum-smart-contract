package common

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
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
	var u BaseStateBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return DecorateError(err, ErrDecodeBson, *s)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return DecorateError(err, ErrDecodeBson, *s)
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
		return DecorateError(err, ErrDecodeBson, *s)
	default:
		s.v = i
	}

	return nil
}
