package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type KeyUpdaterFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	TG base.Address `json:"target"`
	KS AccountKeys  `json:"keys"`
	CR CurrencyID   `json:"currency"`
}

func (fact KeyUpdaterFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(KeyUpdaterFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		TG:                    fact.target,
		KS:                    fact.keys,
		CR:                    fact.currency,
	})
}

type KeyUpdaterFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	TG string          `json:"target"`
	KS json.RawMessage `json:"keys"`
	CR string          `json:"currency"`
}

func (fact *KeyUpdaterFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of KeyUpdaterFact")

	var uf KeyUpdaterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.TG, uf.KS, uf.CR)
}

func (op *KeyUpdater) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
