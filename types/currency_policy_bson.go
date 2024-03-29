package types

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (po CurrencyPolicy) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":                   po.Hint().String(),
			"new_account_min_balance": po.newAccountMinBalance.String(),
			"feeer":                   po.feeer,
		},
	)
}

type CurrencyPolicyBSONUnmarshaler struct {
	Hint          string   `bson:"_hint"`
	NewAccountMin string   `bson:"new_account_min_balance"`
	Feeer         bson.Raw `bson:"feeer"`
}

func (po *CurrencyPolicy) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("Decode bson of CurrencyPolicy")

	var upo CurrencyPolicyBSONUnmarshaler
	if err := enc.Unmarshal(b, &upo); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(upo.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return po.unpack(enc, ht, upo.NewAccountMin, upo.Feeer)
}
