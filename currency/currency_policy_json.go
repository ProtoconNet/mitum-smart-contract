package currency

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type CurrencyPolicyJSONMarshaler struct {
	hint.BaseHinter
	MN Big   `json:"new_account_min_balance"`
	FE Feeer `json:"feeer"`
}

func (po CurrencyPolicy) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyPolicyJSONMarshaler{
		BaseHinter: po.BaseHinter,
		MN:         po.newAccountMinBalance,
		FE:         po.feeer,
	})
}

type CurrencyPolicyJSONUnmarshaler struct {
	MN Big             `json:"new_account_min_balance"`
	FE json.RawMessage `json:"feeer"`
}

func (po *CurrencyPolicy) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of CurrencyPolicy")

	var upo CurrencyPolicyJSONUnmarshaler
	if err := enc.Unmarshal(b, &upo); err != nil {
		return e(err, "")
	}

	po.newAccountMinBalance = upo.MN

	var feeer Feeer
	err := encoder.Decode(enc, upo.FE, &feeer)
	if err != nil {
		return errors.WithMessage(err, "failed to decode feeer")
	}
	po.feeer = feeer

	return nil
}
