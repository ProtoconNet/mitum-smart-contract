package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type FeeOperationFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	AM []Amount `json:"amounts"`
}

func (fact FeeOperationFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FeeOperationFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		AM:                    fact.amounts,
	})
}

type FeeOperationFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	AM json.RawMessage `json:"amounts"`
}

func (fact *FeeOperationFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode FeeOperationFact")

	var uft FeeOperationFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uft); err != nil {
		return e(err, "")
	}

	return fact.unpack(enc, uft)
}
