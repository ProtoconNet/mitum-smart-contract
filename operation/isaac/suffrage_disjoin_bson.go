package isaacoperation

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact SuffrageDisjoinFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": fact.Hint().String(),
			"node":  fact.node,
			"start": fact.start,
			"hash":  fact.BaseFact.Hash().String(),
			"token": fact.BaseFact.Token(),
		},
	)
}

type SuffrageDisjoinFactBSONUnMarshaler struct {
	Hint  string      `bson:"_hint"`
	Node  string      `bson:"node"`
	Start base.Height `bson:"start"`
}

func (fact *SuffrageDisjoinFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of SuffrageDisjoinFact")

	var u common.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf SuffrageDisjoinFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unpack(enc, uf.Node, uf.Start)
}

func (op *SuffrageDisjoin) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of SuffrageDisjoin")
	var ubo common.BaseNodeOperation

	err := ubo.DecodeBSON(b, enc)
	if err != nil {
		return e.Wrap(err)
	}

	op.BaseNodeOperation = ubo

	return nil
}
