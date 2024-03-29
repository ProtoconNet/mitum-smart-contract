package extension // nolint: dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact WithdrawFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":  fact.Hint().String(),
			"sender": fact.sender,
			"items":  fact.items,
			"hash":   fact.BaseFact.Hash().String(),
			"token":  fact.BaseFact.Token(),
		},
	)
}

type WithdrawFactBSONUnmarshaler struct {
	Hint   string   `bson:"_hint"`
	Sender string   `bson:"sender"`
	Items  bson.Raw `bson:"items"`
}

func (fact *WithdrawFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {

	var ubf common.BaseFactBSONUnmarshaler
	err := enc.Unmarshal(b, &ubf)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	h := valuehash.NewBytesFromString(ubf.Hash)

	fact.BaseFact.SetHash(h)
	err = fact.BaseFact.SetToken(ubf.Token)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	var uf WithdrawFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(enc, uf.Sender, uf.Items); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op Withdraw) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(op.BaseOperation)
}

func (op *Withdraw) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {

	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
