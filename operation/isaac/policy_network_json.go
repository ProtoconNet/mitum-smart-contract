package isaacoperation

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type GenesisNetworkPolicyFactJSONMarshaler struct {
	Policy base.NetworkPolicy `json:"policy"`
	base.BaseFactJSONMarshaler
}

func (fact GenesisNetworkPolicyFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(GenesisNetworkPolicyFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Policy:                fact.policy,
	})
}

type GenesisNetworkPolicyFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Policy json.RawMessage `json:"policy"`
}

func (fact *GenesisNetworkPolicyFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode GenesisNetworkPolicyFact")

	var u GenesisNetworkPolicyFactJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}
	fact.BaseFact.SetJSONUnmarshaler(u.BaseFactJSONUnmarshaler)

	if err := encoder.Decode(enc, u.Policy, &fact.policy); err != nil {
		return e.Wrap(err)
	}

	return nil
}
