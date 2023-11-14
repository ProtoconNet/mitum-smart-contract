package types

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type AmountJSONMarshaler struct {
	AmountBig string     `json:"amount"`
	Currency  CurrencyID `json:"currency"`
	hint.BaseHinter
}

func (am Amount) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AmountJSONMarshaler{
		BaseHinter: am.BaseHinter,
		AmountBig:  am.big.String(),
		Currency:   am.cid,
	})
}

type AmountJSONUnmarshaler struct {
	AmountBig string    `json:"amount"`
	Currency  string    `json:"currency"`
	Hint      hint.Hint `json:"_hint"`
}

func (am *Amount) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of Amount")

	var uam AmountJSONUnmarshaler
	if err := enc.Unmarshal(b, &uam); err != nil {
		return e.Wrap(err)
	}

	am.BaseHinter = hint.NewBaseHinter(uam.Hint)

	return am.unpack(enc, uam.Currency, uam.AmountBig)
}
