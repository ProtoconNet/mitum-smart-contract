package contract

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact CallContractFact) MarshalBSON() ([]byte, error) {
	m := bson.M{
		"_hint":    fact.Hint().String(),
		"hash":     fact.BaseFact.Hash().String(),
		"token":    fact.BaseFact.Token(),
		"sender":   fact.sender.String(),
		"contract": fact.contract.String(),
		"currency": fact.currency,
	}
	if len(fact.items) == 1 {
		m["call_data"] = fact.items[0].rebuildLegacyCallData()
	} else {
		m["items"] = fact.items
	}

	return bsonenc.Marshal(
		m,
	)
}

type CallContractFactBSONUnmarshaler struct {
	Hint     string   `bson:"_hint"`
	Sender   string   `bson:"sender"`
	Contract string   `bson:"contract"`
	CallData bson.Raw `bson:"call_data"`
	Items    bson.Raw `bson:"items"`
	Currency string   `bson:"currency"`
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

	switch {
	case len(uf.CallData) > 0 && len(uf.Items) > 0:
		return common.DecorateError(
			common.ErrValueInvalid.Errorf("call_data and items cannot both be set"),
			common.ErrDecodeBson,
			*fact,
		)
	case len(uf.Items) > 0:
		items, err := decodeCallContractItemsBSONRaw(uf.Items)
		if err != nil {
			return common.DecorateError(err, common.ErrDecodeBson, *fact)
		}
		if err := fact.unpackItems(enc, uf.Sender, uf.Contract, items, uf.Currency); err != nil {
			return common.DecorateError(err, common.ErrDecodeBson, *fact)
		}
	default:
		var callData map[string]string
		if len(uf.CallData) > 0 {
			if err := bson.Unmarshal(uf.CallData, &callData); err != nil {
				return common.DecorateError(err, common.ErrDecodeBson, *fact)
			}
		}
		if err := fact.unpackLegacy(enc, uf.Sender, uf.Contract, callData, uf.Currency); err != nil {
			return common.DecorateError(err, common.ErrDecodeBson, *fact)
		}
	}

	return nil
}

func decodeCallContractItemsBSONRaw(raw bson.Raw) ([]CallContractItem, error) {
	values, err := raw.Values()
	if err != nil {
		return nil, err
	}

	items := make([]CallContractItem, len(values))
	for i := range values {
		if err := bson.Unmarshal(values[i].Value, &items[i]); err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (it CallContractItem) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"function":  it.function,
			"call_data": normalizeStringMap(it.callData),
		},
	)
}

type CallContractItemBSONUnmarshaler struct {
	Function string            `bson:"function"`
	CallData map[string]string `bson:"call_data"`
}

func (it *CallContractItem) DecodeBSON(b []byte, _ *bsonenc.Encoder) error {
	var u CallContractItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	*it = NewCallContractItem(u.Function, u.CallData)

	return nil
}

func (it *CallContractItem) UnmarshalBSON(b []byte) error {
	var u CallContractItemBSONUnmarshaler
	if err := bson.Unmarshal(b, &u); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *it)
	}

	*it = NewCallContractItem(u.Function, u.CallData)

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
