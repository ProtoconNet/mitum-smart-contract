package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type KeyUpdaterFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Target   base.Address `json:"target"`
	Keys     AccountKeys  `json:"keys"`
	Currency CurrencyID   `json:"currency"`
}

func (fact KeyUpdaterFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeyUpdaterFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Target:                fact.target,
		Keys:                  fact.keys,
		Currency:              fact.currency,
	})
}

type KeyUpdaterFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	Target   string          `json:"target"`
	Keys     json.RawMessage `json:"keys"`
	Currency string          `json:"currency"`
}

func (fact *KeyUpdaterFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of KeyUpdaterFact")

	var uf KeyUpdaterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.Target, uf.Keys, uf.Currency)
}

type keyUpdaterMarshaler struct {
	BaseOperationJSONMarshaler
}

func (op *KeyUpdater) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode KeyUpdater")

	var ubo BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo

	return nil
}
