package base // nolint: dupl, revive

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v2/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (cs ContractAccount) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    cs.Hint().String(),
			"isactive": cs.isActive,
			"owner":    cs.owner,
		},
	)
}

type ContractAccountBSONUnmarshaler struct {
	Hint     string `json:"_hint"`
	IsActive bool   `bson:"isactive"`
	Owner    string `bson:"owner"`
}

func (cs *ContractAccount) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of ContractAccount")

	var ucs ContractAccountBSONUnmarshaler
	if err := bsonenc.Unmarshal(b, &ucs); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(ucs.Hint)
	if err != nil {
		return e(err, "")
	}

	return cs.unpack(enc, ht, ucs.IsActive, ucs.Owner)
}
