package currency

import (
	"encoding/json"
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type CreateAccountsItemJSONMarshaler struct {
	hint.BaseHinter
	Keys     base2.AccountKeys `json:"keys"`
	Amounts  []base2.Amount    `json:"amounts"`
	AddrType hint.Type         `json:"addrtype"`
}

func (it BaseCreateAccountsItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateAccountsItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Keys:       it.keys,
		Amounts:    it.amounts,
		AddrType:   it.addressType,
	})
}

type CreateAccountsItemJSONUnMarshaler struct {
	Hint     hint.Hint       `json:"_hint"`
	Keys     json.RawMessage `json:"keys"`
	Amounts  json.RawMessage `json:"amounts"`
	AddrType string          `json:"addrtype"`
}

func (it *BaseCreateAccountsItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of BaseCreateAccountsItem")

	var uit CreateAccountsItemJSONUnMarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.Hint, uit.Keys, uit.Amounts, uit.AddrType)
}
