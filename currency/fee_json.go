package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type FeeOperationFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Amounts []Amount `json:"amounts"`
}

func (fact FeeOperationFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(FeeOperationFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Amounts:               fact.amounts,
	})
}

type FeeOperationFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	AM json.RawMessage `json:"amounts"`
}

func (fact *FeeOperationFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of FeeOperationFact")

	var uf FeeOperationFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.AM)
}
