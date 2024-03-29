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
	var uf UpdateKeyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	if err := fact.unpack(enc, uf.Target, uf.Keys, uf.Currency); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	return nil
}

func (op UpdateKey) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseOperationMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateKey) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseOperation = ubo

	return nil
}
