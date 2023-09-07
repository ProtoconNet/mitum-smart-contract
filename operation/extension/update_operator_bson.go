package extension // nolint: dupl

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

func (fact UpdateOperatorFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     fact.Hint().String(),
			"sender":    fact.sender,
			"contract":  fact.contract,
			"operators": fact.operators,
			"currency":  fact.currency,
			"hash":      fact.BaseFact.Hash().String(),
			"token":     fact.BaseFact.Token(),
		},
	)
}

type UpdateOperatorFactBSONUnmarshaler struct {
	Hint      string   `bson:"_hint"`
	Sender    string   `bson:"sender"`
	Contract  string   `bson:"contract"`
	Operators []string `bson:"operators"`
	Currency  string   `bson:"currency"`
}

func (fact *UpdateOperatorFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of UpdateOperatorFact")

	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf UpdateOperatorFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unpack(enc, uf.Sender, uf.Contract, uf.Operators, uf.Currency)
}

func (op UpdateOperator) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		})
}

func (op *UpdateOperator) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of UpdateOperator")

	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
