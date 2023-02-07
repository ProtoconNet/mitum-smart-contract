package currency

import (
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact GenesisCurrenciesFact) MarshalBSON() ([]byte, error) {
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

type GenesisCurrenciesFactBSONUnMarshaler struct {
	HT string   `bson:"_hint"`
	GK string   `bson:"genesis_node_key"`
	KS bson.Raw `bson:"keys"`
	CS bson.Raw `bson:"currencies"`
}

func (fact *GenesisCurrenciesFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of GenesisCurrenciesFact")

	var u BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf GenesisCurrenciesFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uf.HT)
	if err != nil {
		return e(err, "")
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unpack(enc, uf.GK, uf.KS, uf.CS)
}

func (op GenesisCurrencies) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(op.BaseOperation)
}

func (op *GenesisCurrencies) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of GenesisCurrencies")
	var ubo BaseOperation

	err := ubo.DecodeBSON(b, enc)
	if err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo

	return nil
}
