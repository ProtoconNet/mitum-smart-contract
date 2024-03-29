package currency // nolint:dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseCreateAccountItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   it.Hint().String(),
			"keys":    it.keys,
			"amounts": it.amounts,
		},
	)
}

type CreateAccountItemBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Keys   bson.Raw `bson:"keys"`
	Amount bson.Raw `bson:"amounts"`
}

func (it *BaseCreateAccountItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit CreateAccountItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	if err := it.unpack(enc, ht, uit.Keys, uit.Amount); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}
	return nil
}
