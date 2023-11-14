package common

import (
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var NodeHint = hint.MustNewHint("currency-node-v0.0.1")

type BaseNode struct {
	util.IsValider
	addr base.Address
	pub  base.Publickey
	hint.BaseHinter
}

func NewBaseNode(ht hint.Hint, pub base.Publickey, addr base.Address) BaseNode {
	return BaseNode{
		BaseHinter: hint.NewBaseHinter(ht),
		addr:       addr,
		pub:        pub,
	}
}

func (n BaseNode) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, n.addr, n.pub); err != nil {
		return errors.Wrap(err, "invalid RemoteNode")
	}

	return nil
}

func (n BaseNode) Address() base.Address {
	return n.addr
}

func (n BaseNode) Publickey() base.Publickey {
	return n.pub
}

func (n BaseNode) HashBytes() []byte {
	return util.ConcatByters(n.addr, n.pub)
}

type BaseNodeJSONMarshaler struct {
	Address   base.Address   `json:"address"`
	Publickey base.Publickey `json:"publickey"`
}

func (n BaseNode) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(struct {
		BaseNodeJSONMarshaler
		hint.BaseHinter
	}{
		BaseHinter: n.BaseHinter,
		BaseNodeJSONMarshaler: BaseNodeJSONMarshaler{
			Address:   n.addr,
			Publickey: n.pub,
		},
	})
}

type BaseNodeJSONUnmarshaler struct {
	Address   string `json:"address"`
	Publickey string `json:"publickey"`
}

func (n *BaseNode) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode BaseNode")

	var u BaseNodeJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	switch i, err := base.DecodeAddress(u.Address, enc); {
	case err != nil:
		return e.WithMessage(err, "decode node address")
	default:
		n.addr = i
	}

	switch i, err := base.DecodePublickeyFromString(u.Publickey, enc); {
	case err != nil:
		return e.WithMessage(err, "node publickey")
	default:
		n.pub = i
	}

	return nil
}

func (n BaseNode) MarshalBSON() ([]byte, error) {
	return bsonenc.Marshal(
		bson.M{
			"_hint":     n.Hint().String(),
			"address":   n.addr.String(),
			"publickey": n.pub.String(),
		},
	)
}

type BaseNodeBSONUnMarshaler struct {
	Hint      string `bson:"_hint"`
	Address   string `bson:"address"`
	Publickey string `bson:"publickey"`
}

func (n *BaseNode) DecodeBSON(b []byte, enc *bsonenc.Encoder) error {
	e := util.StringError("decode bson of BaseNode")

	var u BaseNodeBSONUnMarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e.Wrap(err)
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e.Wrap(err)
	}
	n.BaseHinter = hint.NewBaseHinter(ht)

	switch i, err := base.DecodeAddress(u.Address, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		n.addr = i
	}

	switch p, err := base.DecodePublickeyFromString(u.Publickey, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		n.pub = p
	}

	return nil
}
