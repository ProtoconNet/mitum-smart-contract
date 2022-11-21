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

	var ufact CurrencyRegisterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &ufact); err != nil {
		return e(err, "")
	}

	return fact.unpack(enc, ufact)
}

func (op *CurrencyRegister) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo base.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
