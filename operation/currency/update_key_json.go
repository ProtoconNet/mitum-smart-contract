package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type UpdateKeyFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Target   base.Address      `json:"target"`
	Keys     types.AccountKeys `json:"keys"`
	Currency types.CurrencyID  `json:"currency"`
}

func (fact UpdateKeyFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateKeyFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Target:                fact.target,
		Keys:                  fact.keys,
		Currency:              fact.currency,
	})
}

type UpdateKeyFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Target   string          `json:"target"`
	Keys     json.RawMessage `json:"keys"`
	Currency string          `json:"currency"`
}

func (fact *UpdateKeyFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of UpdateKeyFact")

	var uf UpdateKeyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Target, uf.Keys, uf.Currency)
}

func (op UpdateKey) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateKey) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode UpdateKey")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
