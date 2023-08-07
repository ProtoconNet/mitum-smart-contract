package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type CreateAccountsFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender base.Address         `json:"sender"`
	Items  []CreateAccountsItem `json:"items"`
}

func (fact CreateAccountsFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateAccountsFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Items:                 fact.items,
	})
}

type CreateAccountsFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender string          `json:"sender"`
	Items  json.RawMessage `json:"items"`
}

func (fact *CreateAccountsFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of CreateAccountsFact")

	var uf CreateAccountsFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)
	return fact.unpack(enc, uf.Sender, uf.Items)
}

type createAccountsMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op CreateAccounts) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(createAccountsMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CreateAccounts) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode CreateAccounts")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
