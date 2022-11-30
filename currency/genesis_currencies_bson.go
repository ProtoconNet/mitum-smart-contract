package currency

import (
	"github.com/pkg/errors"
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
	e := util.StringErrorFunc("failed to decode GenesisCurrenciesFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return err
	}

	fact.BaseFact = ubf

	var ugcf GenesisCurrenciesFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &ugcf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(ugcf.HT)

	switch pk, err := base.DecodePublickeyFromString(ugcf.GK, enc); {
	case err != nil:
		return err
	default:
		fact.genesisNodeKey = pk
	}

	var keys AccountKeys
	hinter, err := enc.Decode(ugcf.KS)
	if err != nil {
		return err
	} else if k, ok := hinter.(AccountKeys); !ok {
		return errors.Errorf("not Keys: %T", hinter)
	} else {
		keys = k
	}

	fact.keys = keys

	hcs, err := enc.DecodeSlice(ugcf.CS)
	if err != nil {
		return err
	}

	cs := make([]CurrencyDesign, len(hcs))
	for i := range hcs {
		j, ok := hcs[i].(CurrencyDesign)
		if !ok {
			return util.ErrWrongType.Errorf("expected CurrencyDesign, not %T", hcs[i])
		}

		cs[i] = j
	}
	fact.cs = cs

	return nil
}

func (op *GenesisCurrencies) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
