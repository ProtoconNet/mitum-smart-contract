package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
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

func (fact *UpdateKeyFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of UpdateKeyFact")

	var uf UpdateKeyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Target, uf.Keys, uf.Currency)
}

type updateKeyMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op UpdateKey) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(updateKeyMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *UpdateKey) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode UpdateKey")

	var ubo common.BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e.Wrap(err)
	}

	op.BaseOperation = ubo

	return nil
}
