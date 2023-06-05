package currency

import (
	"encoding/json"
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type SuffrageInflationFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Items []SuffrageInflationItem `json:"items"`
}

func (fact SuffrageInflationFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SuffrageInflationFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Items:                 fact.items,
	})
}

type SuffrageInflationFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Items []json.RawMessage `json:"items"`
}

func (fact *SuffrageInflationFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of SuffrageInflationFact")

	var uf SuffrageInflationFactJSONUnmarshaler

	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	items := make([]SuffrageInflationItem, len(uf.Items))
	for i := range uf.Items {
		item := SuffrageInflationItem{}
		if err := item.DecodeJSON(uf.Items[i], enc); err != nil {
			return e(err, "")
		}
		items[i] = item
	}

	fact.items = items

	return nil
}

func (op *SuffrageInflation) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo base2.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
