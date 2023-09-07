package currency // nolint:dupl

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateAccountItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    it.Hint().String(),
			"keys":     it.keys,
			"amounts":  it.amounts,
			"addrtype": it.addressType,
		},
	)
}

type CreateAccountItemBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Keys     bson.Raw `bson:"keys"`
	Amount   bson.Raw `bson:"amounts"`
	AddrType string   `bson:"addrtype"`
}

func (it *BaseCreateAccountItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of BaseCreateAccountItem")

	var uit CreateAccountItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, ht, uit.Keys, uit.Amount, uit.AddrType)
}
