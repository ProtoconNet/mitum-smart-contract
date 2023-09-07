package isaacoperation

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (l FixedSuffrageCandidateLimiterRule) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": l.Hint().String(),
			"limit": l.limit,
		},
	)
}

type FixedSuffrageCandidateLimiterRuleBSONUnMarshaler struct {
	Hint  string `bson:"_hint"`
	Limit uint64 `bson:"limit"`
}

func (l *FixedSuffrageCandidateLimiterRule) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of FixedSuffrageCandidateLimiterRule")

	var u FixedSuffrageCandidateLimiterRuleBSONUnMarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	l.BaseHinter = hint.NewBaseHinter(ht)

	l.limit = u.Limit

	return nil
}

func (l MajoritySuffrageCandidateLimiterRule) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": l.Hint().String(),
			"ratio": l.ratio,
			"max":   l.max,
			"min":   l.min,
		},
	)
}

type MajoritySuffrageCandidateLimiterRuleBSONUnMarshaler struct {
	Hint  string  `bson:"_hint"`
	Ratio float64 `bson:"ratio"`
	Max   uint64  `bson:"max"`
	Min   uint64  `bson:"min"`
}

func (l *MajoritySuffrageCandidateLimiterRule) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of MajoritySuffrageCandidateLimiterRule")

	var u MajoritySuffrageCandidateLimiterRuleBSONUnMarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	l.BaseHinter = hint.NewBaseHinter(ht)

	l.ratio = u.Ratio
	l.max = u.Max
	l.min = u.Min

	return nil
}
