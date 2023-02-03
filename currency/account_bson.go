package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ac Account) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":   ac.Hint().String(),
			"hash":    ac.h,
			"address": ac.address,
			"keys":    ac.keys,
		},
	)
}

type AccountBSONUnmarshaler struct {
	HT string          `bson:"_hint"`
	H  valuehash.Bytes `bson:"hash"`
	AD string          `bson:"address"`
	KS bson.Raw        `bson:"keys"`
}

func (ac *Account) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of Account")

	var uac AccountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uac.HT)
	if err != nil {
		return e(err, "")
	}

	ac.h = valuehash.NewHashFromBytes(uac.H)

	ac.BaseHinter = hint.NewBaseHinter(ht)
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
		v, ok := k.(AccountKeys)
		if !ok {
			return util.ErrWrongType.Errorf("expected BaseAccountKeys, not %T", k)
		}
		ac.keys = v
	}

	return nil
}
