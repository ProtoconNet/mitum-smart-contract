package extension

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type ContractAccountStateValueJSONMarshaler struct {
	hint.BaseHinter
	ContractAccount types.ContractAccount `json:"contractaccount"`
}

func (s ContractAccountStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(ContractAccountStateValueJSONMarshaler{
		BaseHinter:      s.BaseHinter,
		ContractAccount: s.account,
	})
}

type ContractAccountStateValueJSONUnmarshaler struct {
	Hint            hint.Hint       `json:"_hint"`
	ContractAccount json.RawMessage `json:"contractaccount"`
}

func (s *ContractAccountStateValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of ContractAccountStateValue")

	var u ContractAccountStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	s.BaseHinter = hint.NewBaseHinter(u.Hint)

	var ca types.ContractAccount
	if err := ca.DecodeJSON(u.ContractAccount, enc); err != nil {
		return e.Wrap(err)
	}
	s.account = ca

	return nil
}
