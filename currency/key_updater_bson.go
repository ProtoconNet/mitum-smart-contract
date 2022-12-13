package currency // nolint: dupl

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
)

func (fact KeyUpdaterFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bsonenc.MergeBSONM(
			bsonenc.NewHintedDoc(fact.Hint()),
			bson.M{
				"target":   fact.target,
				"keys":     fact.keys,
				"currency": fact.currency,
			},
			fact.BaseFact.BSONM(),
		))
}

type KeyUpdaterFactBSONUnmarshaler struct {
	HT hint.Hint `bson:"_hint"`
	TG string    `bson:"target"`
	KS bson.Raw  `bson:"keys"`
	CR string    `bson:"currency"`
}

func (fact *KeyUpdaterFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of KeyUpdaterFact")

	var ubf base.BaseFact
	if err := ubf.DecodeBSON(b, enc); err != nil {
		return err
	}

	fact.BaseFact = ubf

	var ukuf KeyUpdaterFactBSONUnmarshaler
	if err := bson.Unmarshal(b, &ukuf); err != nil {
		return e(err, "")
	}

	fact.BaseHinter = hint.NewBaseHinter(ukuf.HT)
	switch a, err := base.DecodeAddress(ukuf.TG, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.target = a
	}

	if hinter, err := enc.Decode(ukuf.KS); err != nil {
		return err
	} else if k, ok := hinter.(AccountKeys); !ok {
		return e(errors.Errorf("expected AccountsKeys not %T,", hinter), "")
	} else {
		fact.keys = k
	}

	fact.currency = CurrencyID(ukuf.CR)

	return nil
}

func (op *KeyUpdater) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeBSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
