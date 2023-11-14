package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type CreateAccountFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender base.Address        `json:"sender"`
	Items  []CreateAccountItem `json:"items"`
}

func (fact CreateAccountFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateAccountFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Items:                 fact.items,
	})
}

type CreateAccountFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender string          `json:"sender"`
	Items  json.RawMessage `json:"items"`
}

func (fact *CreateAccountFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of CreateAccountFact")

	var uf CreateAccountFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)
	return fact.unpack(enc, uf.Sender, uf.Items)
}

type BaseOperationMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op CreateAccount) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CreateAccount) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode CreateAccount")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
