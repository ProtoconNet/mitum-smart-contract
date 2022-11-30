package digest

import (
	"encoding/json"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type AccountValueJSONMarshaler struct {
	hint.BaseHinter
	currency.AccountJSONMarshaler
	BL []currency.Amount `json:"balance,omitempty"`
	HT base.Height       `json:"height"`
}

func (va AccountValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AccountValueJSONMarshaler{
		BaseHinter:           va.BaseHinter,
		AccountJSONMarshaler: va.ac.EncodeJSON(),
		BL:                   va.balance,
		HT:                   va.height,
	})
}

type AccountValueJSONUnmarshaler struct {
	HT hint.Hint
	BL json.RawMessage `json:"balance"`
	H  base.Height     `json:"height"`
}

func (va *AccountValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var uva AccountValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	ac := new(currency.Account)
	if err := va.unpack(enc, uva.HT, nil, uva.BL, uva.H); err != nil {
		return err
	} else if err := ac.DecodeJSON(b, enc); err != nil {
		return err
	} else {
		va.ac = *ac

		return nil
	}
}
