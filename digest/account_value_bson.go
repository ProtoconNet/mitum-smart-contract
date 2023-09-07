package digest

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (va AccountValue) MarshalBSON() ([]byte, error) {
	//if &va.contractAccountStatus == nil {
	//	return bsonenc.Marshal(bsonenc.MergeBSONM(
	//		bson.M{
	//			"_hint":   va.Hint().String(),
	//			"ac":      va.ac,
	//			"balance": va.balance,
	//			"height":  va.height,
	//		},
	//	))
	//}

	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bson.M{
			"_hint":                   va.Hint().String(),
			"ac":                      va.ac,
			"balance":                 va.balance,
			"height":                  va.height,
			"contract_account_status": va.contractAccountStatus,
		},
	))
}

type AccountValueBSONUnmarshaler struct {
	Hint                  string      `bson:"_hint"`
	Account               bson.Raw    `bson:"ac"`
	Balance               bson.Raw    `bson:"balance"`
	Height                base.Height `bson:"height"`
	ContractAccountStatus bson.Raw    `bson:"contract_account_status"`
}

func (va *AccountValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of AccountValue")

	var uva AccountValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uva.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return va.unpack(enc, ht, uva.Account, uva.Balance, uva.Height, uva.ContractAccountStatus)
}
