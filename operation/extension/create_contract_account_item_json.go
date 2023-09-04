package extension

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type CreateContractAccountItemJSONMarshaler struct {
	hint.BaseHinter
	Keys     types.AccountKeys `json:"keys"`
	Amounts  []types.Amount    `json:"amounts"`
	AddrType hint.Type         `json:"addrtype"`
}

func (it BaseCreateContractAccountItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateContractAccountItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Keys:       it.keys,
		Amounts:    it.amounts,
		AddrType:   it.addressType,
	})
}

type CreateContractAccountItemJSONUnMarshaler struct {
	Hint     hint.Hint       `json:"_hint"`
	Keys     json.RawMessage `json:"keys"`
	Amounts  json.RawMessage `json:"amounts"`
	AddrType string          `json:"addrtype"`
}

func (it *BaseCreateContractAccountItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of BaseCreateContractAccountItem")

	var uit CreateContractAccountItemJSONUnMarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, uit.Hint, uit.Keys, uit.Amounts, uit.AddrType)
}
