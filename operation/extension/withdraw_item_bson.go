package extension // nolint:dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (it BaseWithdrawItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   it.Hint().String(),
			"target":  it.target,
			"amounts": it.amounts,
		},
	)
}

type BaseWithdrawItemBSONUnmarshaler struct {
	Hint    string   `bson:"_hint"`
	Target  string   `bson:"target"`
	Amounts bson.Raw `bson:"amounts"`
}

func (it *BaseWithdrawItem) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var uit BaseWithdrawItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	ht, err := hint.ParseHint(uit.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	if err := it.unpack(enc, ht, uit.Target, uit.Amounts); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	return nil
}
