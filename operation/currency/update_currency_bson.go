package currency // nolint: dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact UpdateCurrencyFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    fact.Hint().String(),
			"currency": fact.currency,
			"policy":   fact.policy,
			"hash":     fact.BaseFact.Hash().String(),
			"token":    fact.BaseFact.Token(),
		},
	)
}

type UpdateCurrencyFactBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Currency string   `bson:"currency"`
	Policy   bson.Raw `bson:"policy"`
}

func (fact *UpdateCurrencyFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of UpdateCurrencyFact")

	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf UpdateCurrencyFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unpack(enc, uf.Currency, uf.Policy)
}

func (op UpdateCurrency) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		})
}

func (op *UpdateCurrency) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of UpdateCurrency")

	var ubo common.BaseNodeOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseNodeOperation = ubo

	return nil
}
