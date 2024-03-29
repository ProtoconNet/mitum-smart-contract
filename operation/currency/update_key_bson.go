package currency // nolint: dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact UpdateKeyFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    fact.Hint().String(),
			"target":   fact.target,
			"keys":     fact.keys,
			"currency": fact.currency,
			"hash":     fact.BaseFact.Hash().String(),
			"token":    fact.BaseFact.Token(),
		},
	)
}

type UpdateKeyFactBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Target   string   `bson:"target"`
	Keys     bson.Raw `bson:"keys"`
	Currency string   `bson:"currency"`
}

func (fact *UpdateKeyFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	h := valuehash.NewBytesFromString(u.Hash)

	fact.BaseFact.SetHash(h)
	err = fact.BaseFact.SetToken(u.Token)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	var uf UpdateKeyFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(enc, uf.Target, uf.Keys, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op UpdateKey) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		})
}

func (op *UpdateKey) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
