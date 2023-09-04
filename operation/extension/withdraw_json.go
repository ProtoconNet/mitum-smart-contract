package extension

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type WithdrawFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Sender base.Address   `json:"sender"`
	Items  []WithdrawItem `json:"items"`
}

func (fact WithdrawFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(WithdrawFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Sender:                fact.sender,
		Items:                 fact.items,
	})
}

type WithdrawFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Sender string          `json:"sender"`
	Items  json.RawMessage `json:"items"`
}

func (fact *WithdrawFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of WithdrawFact")

	var uf WithdrawFactJSONUnmarshaler

	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Sender, uf.Items)
}

type withdrawMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op Withdraw) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(withdrawMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *Withdraw) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of Withdraw")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
