package isaacoperation

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"go.mongodb.org/mongo-driver/bson"
)

func (fact SuffrageJoinFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     fact.Hint().String(),
			"candidate": fact.candidate,
			"start":     fact.start,
			"hash":      fact.BaseFact.Hash().String(),
			"token":     fact.BaseFact.Token(),
		},
	)
}

type SuffrageJoinFactBSONUnMarshaler struct {
	Hint      string           `bson:"_hint"`
	Candidate string           `bson:"candidate"`
	Start     mitumbase.Height `bson:"start"`
}

func (fact *SuffrageJoinFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageJoinFact")

	var u base.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf SuffrageJoinFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e(err, "")
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	return fact.unpack(enc, uf.Candidate, uf.Start)
}

func (op *SuffrageJoin) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageJoin")
	var ubo base.BaseNodeOperation

	err := ubo.DecodeBSON(b, enc)
	if err != nil {
		return e(err, "")
	}

	op.BaseNodeOperation = ubo

	return nil
}

func (fact SuffrageGenesisJoinFact) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint": fact.Hint().String(),
			"nodes": fact.nodes,
			"hash":  fact.BaseFact.Hash().String(),
			"token": fact.BaseFact.Token(),
		},
	)
}

type SuffrageGenesisJoinFactBSONUnMarshaler struct {
	Hint  string   `bson:"_hint"`
	Nodes bson.Raw `bson:"nodes"`
}

func (fact *SuffrageGenesisJoinFact) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageGenesisJoinFact")

	var u base.BaseFactBSONUnmarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetHash(valuehash.NewBytesFromString(u.Hash))
	fact.BaseFact.SetToken(u.Token)

	var uf SuffrageGenesisJoinFactBSONUnMarshaler
	if err := bson.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(uf.Hint)
	if err != nil {
		return e(err, "")
	}
	fact.BaseHinter = hint.NewBaseHinter(ht)

	r, err := uf.Nodes.Values()
	if err != nil {
		return err
	}

	nodes := make([]base.BaseNode, len(r))
	for i := range r {
		node := base.BaseNode{}
		if err := node.DecodeBSON(r[i].Value, enc); err != nil {
			return err
		}
		nodes[i] = node
	}

	return nil
}

func (op *SuffrageGenesisJoin) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode bson of SuffrageGenesisJoin")
	var ubo base.BaseOperation

	err := ubo.DecodeBSON(b, enc)
	if err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo

	return nil
}
