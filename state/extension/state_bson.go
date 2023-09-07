package extension

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (c ContractAccountStateValue) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":           c.Hint().String(),
			"contractaccount": c.status,
		},
	)

}

type ContractAccountStateValueBSONUnmarshaler struct {
	Hint            string   `bson:"_hint"`
	ContractAccount bson.Raw `bson:"contractaccount"`
}

func (c *ContractAccountStateValue) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of ContractAccountStateValue")

	var u ContractAccountStateValueBSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	c.BaseHinter = hint.NewBaseHinter(ht)

	var ca types.ContractAccountStatus
	if err := ca.DecodeBSON(u.ContractAccount, enc); err != nil {
		return e.Wrap(err)
	}

	c.status = ca

	return nil
}
