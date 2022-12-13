package currency

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type CurrencyPolicyUpdaterFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	CR CurrencyID     `json:"currency"`
	PO CurrencyPolicy `json:"policy"`
}

func (fact CurrencyPolicyUpdaterFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyPolicyUpdaterFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		CR:                    fact.currency,
		PO:                    fact.policy,
	})
}

type CurrencyPolicyUpdaterFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	CR string          `json:"currency"`
	PO json.RawMessage `json:"policy"`
}

func (fact *CurrencyPolicyUpdaterFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of CurrencyPolicyUpdaterFact")

	var ucpu CurrencyPolicyUpdaterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &ucpu); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(ucpu.BaseFactJSONUnmarshaler)

	fact.currency = CurrencyID(ucpu.CR)

	if hinter, err := enc.Decode(ucpu.PO); err != nil {
		return err
	} else if po, ok := hinter.(CurrencyPolicy); !ok {
		return e(errors.Errorf("expected CurrencyPolicy not %T,", hinter), "")
	} else {
		fact.policy = po
	}

	return nil
}

func (op *CurrencyPolicyUpdater) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo base.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
