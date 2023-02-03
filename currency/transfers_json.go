package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type TransferFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	SD base.Address    `json:"sender"`
	IT []TransfersItem `json:"items"`
}

func (fact TransfersFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransferFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		SD:                    fact.sender,
		IT:                    fact.items,
	})
}

type TransfersFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	SD string          `json:"sender"`
	IT json.RawMessage `json:"items"`
}

func (fact *TransfersFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of TransfersFact")

	var uf TransfersFactJSONUnmarshaler

	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.SD, uf.IT)
}

type transfersMarshaler struct {
	BaseOperationJSONMarshaler
	Memo string `json:memo`
}

func (op Transfers) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(transfersMarshaler{
		BaseOperationJSONMarshaler: op.BaseOperation.JSONMarshaler(),
		Memo:                       op.Memo,
	})
}

func (op *Transfers) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode Transfers")

	var ubo BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return e(err, "")
	}

	var m MemoJSONUnMarshaler
	if err := enc.Unmarshal(b, &m); err != nil {
		return e(err, "")
	}

	op.BaseOperation = ubo
	op.Memo = m.Memo

	return nil
}
