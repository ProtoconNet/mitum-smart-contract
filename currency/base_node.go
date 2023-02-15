package currency

import (
	"github.com/pkg/errors"
	bsonenc "github.com/spikeekips/mitum-currency/digest/util/bson"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"go.mongodb.org/mongo-driver/bson"
)

var NodeHint = hint.MustNewHint("currency-node-v0.0.1")

type BaseNode struct {
	util.IsValider
	addr base.Address
	pub  base.Publickey
	util.DefaultJSONMarshaled
	hint.BaseHinter
}

func NewBaseNode(ht hint.Hint, pub base.Publickey, addr Address) BaseNode {
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

func (n *BaseNode) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode RemoteNode")

	var u BaseNodeJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e(err, "")
	}

	switch i, err := base.DecodeAddress(u.Address, enc); {
	case err != nil:
		return e(err, "failed to decode node address")
	default:
		n.addr = i
	}

	switch i, err := base.DecodePublickeyFromString(u.Publickey, enc); {
	case err != nil:
		return e(err, "failed to decode node publickey")
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
	e := util.StringErrorFunc("failed to decode bson of BaseNode")

	var u BaseNodeBSONUnMarshaler

	err := enc.Unmarshal(b, &u)
	if err != nil {
		return e(err, "")
	}

	ht, err := hint.ParseHint(u.Hint)
	if err != nil {
		return e(err, "")
	}
	n.BaseHinter = hint.NewBaseHinter(ht)

	switch i, err := base.DecodeAddress(u.Address, enc); {
	case err != nil:
		return e(err, "")
	default:
		n.addr = i
	}

	switch p, err := base.DecodePublickeyFromString(u.Publickey, enc); {
	case err != nil:
		return e(err, "")
	default:
		n.pub = p
	}

	return nil
}
