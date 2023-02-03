package currency

import (
	"encoding/json"

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

	var uf CurrencyPolicyUpdaterFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.CR, uf.PO)
}

type currencyPolicyUpdaterMarshaler struct {
	BaseOperationJSONMarshaler
	Memo string `json:memo`
}

func (op CurrencyPolicyUpdater) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(currencyPolicyUpdaterMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
		Memo:                       op.Memo,
	})
}

func (op *CurrencyPolicyUpdater) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CurrencyPolicyUpdater")

	var ubo BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	var m MemoJSONUnMarshaler
	if err := enc.Unmarshal(b, &m); err != nil {
		return e(err, "")
	}

	op.BaseNodeOperation = ubo
	op.Memo = m.Memo

	return nil
}
