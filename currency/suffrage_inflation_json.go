package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type SuffrageInflationFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	IT []SuffrageInflationItem `json:"items"`
}

func (fact SuffrageInflationFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SuffrageInflationFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		IT:                    fact.items,
	})
}

type SuffrageInflationFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	IT []json.RawMessage `json:"items"`
}

func (fact *SuffrageInflationFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of SuffrageInflationFact")

	var uf SuffrageInflationFactJSONUnmarshaler

	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	items := make([]SuffrageInflationItem, len(uf.IT))
	for i := range uf.IT {
		item := SuffrageInflationItem{}
		if err := item.DecodeJSON(uf.IT[i], enc); err != nil {
			return e(err, "")
		}
		items[i] = item
	}

	fact.items = items

	return nil
}

func (op *SuffrageInflation) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo base.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	op.BaseNodeOperation = ubo

	return nil
}
