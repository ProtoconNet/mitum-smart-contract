package extension

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type CreateContractAccountFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Owner base.Address                `json:"sender"`
	Items []CreateContractAccountItem `json:"items"`
}

func (fact CreateContractAccountFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateContractAccountFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Owner:                 fact.sender,
		Items:                 fact.items,
	})
}

type CreateContractAccountFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Owner string          `json:"sender"`
	Items json.RawMessage `json:"items"`
}

func (fact *CreateContractAccountFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of CreateContractAccountFact")

	var uf CreateContractAccountFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Owner, uf.Items)
}

type createContractAccountMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op CreateContractAccount) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(createContractAccountMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CreateContractAccount) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of CreateContractAccount")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
