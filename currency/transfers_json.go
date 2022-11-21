package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type TransferFactJSONPacker struct {
	base.BaseFactJSONMarshaler
	SD base.Address    `json:"sender"`
	IT []TransfersItem `json:"items"`
}

func (fact TransfersFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransferFactJSONPacker{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		SD:                    fact.sender,
		IT:                    fact.items,
	})
}

type TransfersFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	SD string          `json:"sender"`
	IT json.RawMessage `json:"items"`
}

func (fact *TransfersFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode TransfersFact")

	var ufact TransfersFactJSONUnMarshaler

	if err := enc.Unmarshal(b, &ufact); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(ufact.BaseFactJSONUnmarshaler)

	switch a, err := base.DecodeAddress(ufact.SD, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(ufact.IT)
	if err != nil {
		return e(err, "")
	}

	items := make([]TransfersItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(TransfersItem)
		if !ok {
			return util.ErrWrongType.Errorf("expected TransfersItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
