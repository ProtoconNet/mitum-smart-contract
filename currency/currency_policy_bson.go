package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/bson"
)

func (po CurrencyPolicy) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(po.Hint()),
		bson.M{
			"new_account_min_balance": po.newAccountMinBalance,
			"feeer":                   po.feeer,
		}),
	)
}

type CurrencyPolicyBSONUnmarshaler struct {
	MN Big      `bson:"new_account_min_balance"`
	FE bson.Raw `bson:"feeer"`
}

func (po *CurrencyPolicy) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var upo CurrencyPolicyBSONUnmarshaler
	if err := enc.Unmarshal(b, &upo); err != nil {
		return err
	}

	return po.unpack(enc, upo.MN, upo.FE)
}
