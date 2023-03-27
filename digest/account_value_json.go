package digest

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v2/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type AccountValueJSONMarshaler struct {
	hint.BaseHinter
	currency.AccountJSONMarshaler
	Balance []currency.Amount `json:"balance,omitempty"`
	Height  base.Height       `json:"height"`
}

func (va AccountValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AccountValueJSONMarshaler{
		BaseHinter:           va.BaseHinter,
		AccountJSONMarshaler: va.ac.EncodeJSON(),
		Balance:              va.balance,
		Height:               va.height,
	})
}

type AccountValueJSONUnmarshaler struct {
	Hint    hint.Hint
	Balance json.RawMessage `json:"balance"`
	Height  base.Height     `json:"height"`
}

func (va *AccountValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var uva AccountValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	ac := new(currency.Account)
	if err := va.unpack(enc, uva.Hint, nil, uva.Balance, uva.Height); err != nil {
		return err
	} else if err := ac.DecodeJSON(b, enc); err != nil {
		return err
	} else {
		va.ac = *ac

		return nil
	}
}
