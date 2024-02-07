package common

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type BaseFactBSONUnmarshaler struct {
	Hash  string `bson:"hash"`
	Token []byte `bson:"token"`
}

type BaseSignBSONUnmarshaler struct {
	Signer    string         `bson:"signer"`
	Signature base.Signature `bson:"signature"`
	SignedAt  time.Time      `bson:"signed_at"`
}

type BaseOperationBSONUnmarshaler struct {
	Hint  string     `bson:"_hint"`
	Hash  string     `bson:"hash"`
	Fact  bson.Raw   `bson:"fact"`
	Signs []bson.Raw `bson:"signs"`
}

func (op BaseOperation) MarshalBSON() ([]byte, error) {
	var signs bson.A

	for i := range op.signs {
		signs = append(signs, bson.M{
			"signer":    op.signs[i].Signer().String(),
			"signature": op.signs[i].Signature().String(),
			"signed_at": op.signs[i].SignedAt(),
		})
	}

	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": signs,
		},
	)
}

func (op *BaseOperation) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of BaseOperation")

	var u BaseOperationBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	op.BaseHinter = hint.NewBaseHinter(ht)
	op.h = valuehash.NewBytesFromString(u.Hash)

	var fact base.Fact
	if err := encoder.Decode(enc, u.Fact, &fact); err != nil {
		return e.WithMessage(err, "decode fact")
	}

	op.SetFact(fact)

	var signs []base.Sign

	for i := range u.Signs {
		var us BaseSignBSONUnmarshaler
		var pubKey base.Publickey
		var err error
		if err = enc.Unmarshal(u.Signs[i], &us); err != nil {
			return e.Wrap(err)
		}

		if pubKey, err = base.DecodePublickeyFromString(us.Signer, enc); err != nil {
			return e.Wrap(err)
		}

		sign := base.NewBaseSign(pubKey, us.Signature, us.SignedAt)
		signs = append(signs, sign)
	}
	op.signs = signs

	return nil
}

func (op *BaseNodeOperation) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of BaseNodeOperation")

	var u BaseOperationBSONUnmarshaler

	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation.BaseHinter = hint.NewBaseHinter(ht)
	op.BaseOperation.h = valuehash.NewBytesFromString(u.Hash)

	var fact base.Fact
	if err := encoder.Decode(enc, u.Fact, &fact); err != nil {
		return e.WithMessage(err, "decode fact")
	}

	op.BaseOperation.SetFact(fact)

	return nil
}
