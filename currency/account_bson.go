package currency

import (
	"go.mongodb.org/mongo-driver/bson"

	bsonenc "github.com/ProtoconNet/mitum-currency/v2/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
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
	Hint    string          `bson:"_hint"`
	Hash    valuehash.Bytes `bson:"hash"`
	Address string          `bson:"address"`
	Keys    bson.Raw        `bson:"keys"`
}

func (ac *Account) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of Account")

	var uac AccountBSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uac.Hint)
	if err != nil {
		return e(err, "")
	}

	ac.h = valuehash.NewHashFromBytes(uac.Hash)

	ac.BaseHinter = hint.NewBaseHinter(ht)
	switch ad, err := base.DecodeAddress(uac.Address, enc); {
	case err != nil:
		return e(err, "")
	default:
		ac.address = ad
	}

	k, err := enc.Decode(uac.Keys)
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
