package contract

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact CallContractFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     fact.Hint().String(),
			"hash":      fact.BaseFact.Hash().String(),
			"token":     fact.BaseFact.Token(),
			"sender":    fact.sender,
			"contract":  fact.contract,
			"call_data": fact.callData,
			"currency":  fact.currency,
		},
	)
}

type CallContractFactBSONUnmarshaler struct {
	Hint     string            `bson:"_hint"`
	Sender   string            `bson:"sender"`
	Contract string            `bson:"contract"`
	CallData map[string]string `bson:"call_data"`
	Currency string            `bson:"currency"`
}

func (fact *CallContractFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf CallContractFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(enc, uf.Sender, uf.Contract, uf.CallData, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op CallContract) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		},
	)
}

func (op *CallContract) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
