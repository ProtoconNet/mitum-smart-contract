package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type CurrencyRegisterFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Currency types.CurrencyDesign `json:"currency"`
}

func (fact CurrencyRegisterFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyRegisterFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Currency:              fact.currency,
	})
}

type CurrencyRegisterFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Currency json.RawMessage `json:"currency"`
}

func (fact *CurrencyRegisterFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyRegisterFact")

	var uf CurrencyRegisterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Currency)
}

type currencyRegisterMarshaler struct {
	common.BaseOperationJSONMarshaler
}

func (op CurrencyRegister) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(currencyRegisterMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CurrencyRegister) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyRegister")

	var ubo common.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e(err, "")
	}

	op.BaseNodeOperation = ubo

	return nil
}
