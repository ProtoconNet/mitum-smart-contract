package currency

import (
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type AmountJSONMarshaler struct {
	BG string     `json:"amount"`
	CR CurrencyID `json:"currency"`
	hint.BaseHinter
}

func (am Amount) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AmountJSONMarshaler{
		BaseHinter: am.BaseHinter,
		BG:         am.big.String(),
		CR:         am.cid,
	})
}

type AmountJSONUnmarshaler struct {
	BG string    `json:"amount"`
	CR string    `json:"currency"`
	HT hint.Hint `json:"_hint"`
}

func (am *Amount) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of Amount")

	var uam AmountJSONUnmarshaler
	if err := enc.Unmarshal(b, &uam); err != nil {
		return e(err, "")
	}

	am.BaseHinter = hint.NewBaseHinter(uam.HT)

	return am.unpack(enc, uam.CR, uam.BG)
}
