package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type CurrencyPolicyJSONMarshaler struct {
	hint.BaseHinter
	MN string `json:"new_account_min_balance"`
	FE Feeer  `json:"feeer"`
}

func (po CurrencyPolicy) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyPolicyJSONMarshaler{
		BaseHinter: po.BaseHinter,
		MN:         po.newAccountMinBalance.String(),
		FE:         po.feeer,
	})
}

type CurrencyPolicyJSONUnmarshaler struct {
	HT hint.Hint       `json:"_hint"`
	MN string          `json:"new_account_min_balance"`
	FE json.RawMessage `json:"feeer"`
}

func (po *CurrencyPolicy) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of CurrencyPolicy")

	var upo CurrencyPolicyJSONUnmarshaler
	if err := enc.Unmarshal(b, &upo); err != nil {
		return e(err, "")
	}

	return po.unpack(enc, upo.HT, upo.MN, upo.FE)
}
