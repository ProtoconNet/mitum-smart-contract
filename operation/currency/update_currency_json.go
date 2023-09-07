package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type UpdateCurrencyFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Currency types.CurrencyID     `json:"currency"`
	Policy   types.CurrencyPolicy `json:"policy"`
}

func (fact UpdateCurrencyFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(UpdateCurrencyFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Currency:              fact.currency,
		Policy:                fact.policy,
	})
}

type UpdateCurrencyFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Currency string          `json:"currency"`
	Policy   json.RawMessage `json:"policy"`
}

func (fact *UpdateCurrencyFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("decode json of UpdateCurrencyFact")

	var uf UpdateCurrencyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Currency, uf.Policy)
}

func (op UpdateCurrency) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateCurrency) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("decode UpdateCurrency")

	var ubo common.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseNodeOperation = ubo

	return nil
}
