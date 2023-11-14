package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type ContractAccountStatusJSONMarshaler struct {
	hint.BaseHinter
	Owner     base.Address   `json:"owner"`
	IsActive  bool           `json:"is_active"`
	Operators []base.Address `json:"operators"`
}

func (cs ContractAccountStatus) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(ContractAccountStatusJSONMarshaler{
		BaseHinter: cs.BaseHinter,
		Owner:      cs.owner,
		IsActive:   cs.isActive,
		Operators:  cs.operators,
	})
}

type ContractAccountStatusJSONUnmarshaler struct {
	Hint      hint.Hint `json:"_hint"`
	Owner     string    `json:"owner"`
	IsActive  bool      `json:"is_active"`
	Operators []string  `json:"operators"`
}

func (cs *ContractAccountStatus) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of ContractAccountStatus")

	var ucs ContractAccountStatusJSONUnmarshaler
	if err := enc.Unmarshal(b, &ucs); err != nil {
		return e.Wrap(err)
	}

	return cs.unpack(enc, ucs.Hint, ucs.Owner, ucs.IsActive, ucs.Operators)
}
