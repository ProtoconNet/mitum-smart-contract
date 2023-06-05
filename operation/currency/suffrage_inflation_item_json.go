package currency

import (
	"encoding/json"
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"

	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type SuffrageInflationItemJSONMarshaler struct {
	hint.BaseHinter
	Receiver base.Address `json:"receiver"`
	Amount   base2.Amount `json:"amount"`
}

func (it SuffrageInflationItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SuffrageInflationItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Receiver:   it.receiver,
		Amount:     it.amount,
	})
}

type SuffrageInflationItemJSONUnmarshaler struct {
	HT       hint.Hint       `json:"_hint"`
	Receiver string          `json:"receiver"`
	Amount   json.RawMessage `json:"amount"`
}

func (it *SuffrageInflationItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of SuffrageInflationItem")

	var uit SuffrageInflationItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.HT, uit.Receiver, uit.Amount)
}
