package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type ContractAccountJSONMarshaler struct {
	hint.BaseHinter
	IsActive bool         `json:"is_active"`
	Owner    base.Address `json:"owner"`
}

func (cs ContractAccountStatus) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(ContractAccountJSONMarshaler{
		BaseHinter: cs.BaseHinter,
		IsActive:   cs.isActive,
		Owner:      cs.owner,
	})
}

type ContractAccountJSONUnmarshaler struct {
	Hint     hint.Hint `json:"_hint"`
	IsActive bool      `json:"is_active"`
	Owner    string    `json:"owner"`
}

func (ca *ContractAccountStatus) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of ContractAccountStatus")

	var ucs ContractAccountJSONUnmarshaler
	if err := enc.Unmarshal(b, &ucs); err != nil {
		return e.Wrap(err)
	}

	return ca.unpack(enc, ucs.Hint, ucs.IsActive, ucs.Owner)
}
