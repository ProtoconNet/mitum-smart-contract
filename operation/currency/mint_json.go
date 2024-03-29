package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type MintFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	Items []MintItem `json:"items"`
}

func (fact MintFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(MintFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		Items:                 fact.items,
	})
}

type MintFactJSONUnmarshaler struct {
	base.BaseFactJSONUnmarshaler
	Items []json.RawMessage `json:"items"`
}

func (fact *MintFact) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uf MintFactJSONUnmarshaler

	if err := enc.Unmarshal(b, &uf); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *fact)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	items := make([]MintItem, len(uf.Items))
	for i := range uf.Items {
		item := MintItem{}
		if err := item.DecodeJSON(uf.Items[i], enc); err != nil {
			return common.DecorateError(err, common.ErrDecodeJson, *fact)
		}
		items[i] = item
	}

	fact.items = items

	return nil
}

func (op *Mint) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var ubo common.BaseNodeOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *op)
	}

	op.BaseNodeOperation = ubo

	return nil
}
