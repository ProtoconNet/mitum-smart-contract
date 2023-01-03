package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact GenesisCurrenciesFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(fact.Hint()),
		bson.M{
			"genesis_node_key": fact.genesisNodeKey.String(),
			"keys":             fact.keys,
			"currencies":       fact.cs,
		},
		fact.BaseFact.BSONM(),
	))
}

type GenesisCurrenciesFactBSONUnMarshaler struct {
	HT hint.Hint `bson:"_hint"`
	GK string    `bson:"genesis_node_key"`
	KS bson.Raw  `bson:"keys"`
	CS bson.Raw  `bson:"currencies"`
}

func (fact *GenesisCurrenciesFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of GenesisCurrenciesFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return e(err, "")
	}

	fact.BaseFact = ubf

	var uf GenesisCurrenciesFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(uf.HT)

	return fact.unpack(enc, uf.GK, uf.KS, uf.CS)
}

func (op *GenesisCurrencies) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
