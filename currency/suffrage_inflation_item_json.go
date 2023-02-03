package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"

	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type SuffrageInflationItemJSONMarshaler struct {
	hint.BaseHinter
	RC base.Address `json:"receiver"`
	AM Amount       `json:"amount"`
}

func (it SuffrageInflationItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SuffrageInflationItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		RC:         it.receiver,
		AM:         it.amount,
	})
}

type SuffrageInflationItemJSONUnmarshaler struct {
	HT hint.Hint       `json:"_hint"`
	RC string          `json:"receiver"`
	AM json.RawMessage `json:"amount"`
}

func (it *SuffrageInflationItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of SuffrageInflationItem")

	var uit SuffrageInflationItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.HT, uit.RC, uit.AM)
}
