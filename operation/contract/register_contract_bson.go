package contract

import (
	"github.com/imfact-labs/currency-model/common"
	bsonenc "github.com/imfact-labs/currency-model/utils/bsonenc"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (fact RegisterContractFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     fact.Hint().String(),
			"hash":      fact.BaseFact.Hash().String(),
			"token":     fact.BaseFact.Token(),
			"sender":    fact.sender.String(),
			"contract":  fact.contract.String(),
			"code":      fact.code,
			"init_data": fact.callData,
			"currency":  fact.currency,
		},
	)
}

type RegisterModelFactBSONUnmarshaler struct {
	Hint     string            `bson:"_hint"`
	Sender   bson.RawValue     `bson:"sender"`
	Contract bson.RawValue     `bson:"contract"`
	Code     string            `bson:"code"`
	CallData map[string]string `bson:"init_data"`
	Currency string            `bson:"currency"`
}

func (fact *RegisterContractFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf RegisterModelFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	sender, err := decodeRegisterContractAddressBSON(uf.Sender, enc)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	contract, err := decodeRegisterContractAddressBSON(uf.Contract, enc)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	if err := fact.unpack(enc, sender, contract, uf.Code, uf.CallData, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func decodeRegisterContractAddressBSON(raw bson.RawValue, enc *bsonenc.Encoder) (string, error) {
	if raw.Type == bson.TypeString {
		return raw.StringValue(), nil
	}

	doc, ok := raw.DocumentOK()
	if !ok {
		return "", common.ErrValueInvalid.Errorf("expected address string or document, got %s", raw.Type)
	}
	i, err := enc.Decode(doc)
	if err != nil {
		return "", err
	}
	address, ok := i.(interface{ String() string })
	if !ok {
		return "", common.ErrValueInvalid.Errorf("decoded address has unexpected type %T", i)
	}
	return address.String(), nil
}

func (op RegisterContract) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": op.Hint().String(),
			"hash":  op.Hash().String(),
			"fact":  op.Fact(),
			"signs": op.Signs(),
		},
	)
}

func (op *RegisterContract) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
