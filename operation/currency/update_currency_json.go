package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
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

func (fact *UpdateCurrencyFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf UpdateCurrencyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, uf.Currency, uf.Policy); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op UpdateCurrency) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateCurrency) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseNodeOperation = ubo

	return nil
}
