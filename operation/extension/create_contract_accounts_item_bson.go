package extension // nolint:dupl

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v2/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateContractAccountsItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":    it.Hint().String(),
			"keys":     it.keys,
			"amounts":  it.amounts,
			"addrtype": it.addressType,
		},
	)
}

type CreateContractAccountsItemBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Keys     bson.Raw `bson:"keys"`
	Amounts  bson.Raw `bson:"amounts"`
	AddrType string   `bson:"addrtype"`
}

func (it *BaseCreateContractAccountsItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of BaseCreateContractAccountsItem")

	var uit CreateContractAccountsItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return e(err, "")
	}

	return it.unpack(enc, ht, uit.Keys, uit.Amounts, uit.AddrType)
}
