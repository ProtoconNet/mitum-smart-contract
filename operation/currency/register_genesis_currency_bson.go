package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact RegisterGenesisCurrencyFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":            fact.Hint().String(),
			"genesis_node_key": fact.genesisNodeKey.String(),
			"keys":             fact.keys,
			"currencies":       fact.cs,
			"hash":             fact.BaseFact.Hash().String(),
			"token":            fact.BaseFact.Token(),
		},
	)
}

type RegisterGenesisCurrencyFactBSONUnMarshaler struct {
	Hint           string   `bson:"_hint"`
	GenesisNodeKey string   `bson:"genesis_node_key"`
	Keys           bson.Raw `bson:"keys"`
	Currencies     bson.Raw `bson:"currencies"`
}

func (fact *RegisterGenesisCurrencyFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	h := valuehash.NewBytesFromString(u.Hash)

	fact.BaseFact.SetHash(h)
	err = fact.BaseFact.SetToken(u.Token)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	var uf RegisterGenesisCurrencyFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	if err := fact.unpack(enc, uf.GenesisNodeKey, uf.Keys, uf.Currencies); err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *fact)
	}

	return nil
}

func (op RegisterGenesisCurrency) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(op.BaseOperation)
}

func (op *RegisterGenesisCurrency) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo common.BaseOperation

	err := ubo.DecodeBSON(b, enc)
	if err != nil {
		return common.DecorateError(err, common.ErrDecodeBson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
