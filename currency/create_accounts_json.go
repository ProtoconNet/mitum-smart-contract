package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type CreateAccountsFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	SD base.Address         `json:"sender"`
	IT []CreateAccountsItem `json:"items"`
}

func (fact CreateAccountsFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateAccountsFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		SD:                    fact.sender,
		IT:                    fact.items,
	})
}

type CreateAccountsFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	SD string          `json:"sender"`
	IT json.RawMessage `json:"items"`
}

func (fact *CreateAccountsFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of CreateAccountsFact")

	var uf CreateAccountsFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.SD, uf.IT)
}

type createAccountsMarshaler struct {
	BaseOperationJSONMarshaler
	Memo string `json:memo`
}

func (op CreateAccounts) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(createAccountsMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
		Memo:                       op.Memo,
	})
}

func (op *CreateAccounts) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CreateAccounts")

	var ubo BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e(err, "")
	}

	var m MemoJSONUnMarshaler
	if err := enc.Unmarshal(b, &m); err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo
	op.Memo = m.Memo

	return nil
}
