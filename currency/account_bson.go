package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ac Account) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(bsonenc.MergeBSONM(
		bsonenc.NewHintedDoc(ac.Hint()),
		bson.M{
			"hash":    ac.h,
			"address": ac.address,
			"keys":    ac.keys,
		},
	))
}

type AccountBSONUnmarshaler struct {
	HT hint.Hint       `bson:"_hint"`
	H  valuehash.Bytes `bson:"hash"`
	AD string          `bson:"address"`
	KS bson.Raw        `bson:"keys"`
}

func (ac *Account) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal bson of Account")

	var uac AccountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return err
	}

	ac.BaseHinter = hint.NewBaseHinter(uac.HT)

	switch ad, err := base.DecodeAddress(uac.AD, enc); {
	case err != nil:
		return e(err, "")
	default:
		ac.address = ad
	}

	k, err := enc.Decode(uac.KS)
	if err != nil {
		return e(err, "")
	} else if k != nil {
		v, ok := k.(BaseAccountKeys)
		if !ok {
			return util.ErrWrongType.Errorf("expected Keys, not %T", k)
		}
		ac.keys = v
	}

	ac.h = uac.H

	return nil
}
