package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type CurrencyRegisterFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	CR CurrencyDesign `json:"currency"`
}

func (fact CurrencyRegisterFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyRegisterFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		CR:                    fact.currency,
	})
}

type CurrencyRegisterFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	CR json.RawMessage `json:"currency"`
}

func (fact *CurrencyRegisterFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyRegisterFact")

	var uf CurrencyRegisterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.CR)
}

type currencyRegisterMarshaler struct {
	BaseOperationJSONMarshaler
}

func (op CurrencyRegister) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(currencyRegisterMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
	})
}

func (op *CurrencyRegister) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyRegister")

	var ubo BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e(err, "")
	}

	op.BaseNodeOperation = ubo

	return nil
}
