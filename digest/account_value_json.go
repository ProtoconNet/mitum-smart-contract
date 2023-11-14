package digest

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type AccountValueJSONMarshaler struct {
	hint.BaseHinter
	types.AccountJSONMarshaler
	Balance               []types.Amount              `json:"balance,omitempty"`
	Height                base.Height                 `json:"height"`
	ContractAccountStatus types.ContractAccountStatus `json:"contract_account_status"`
}

func (va AccountValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AccountValueJSONMarshaler{
		BaseHinter:            va.BaseHinter,
		AccountJSONMarshaler:  va.ac.EncodeJSON(),
		Balance:               va.balance,
		Height:                va.height,
		ContractAccountStatus: va.contractAccountStatus,
	})
}

type AccountValueJSONUnmarshaler struct {
	Hint                  hint.Hint
	Balance               json.RawMessage `json:"balance"`
	Height                base.Height     `json:"height"`
	ContractAccountStatus json.RawMessage `json:"contract_account_status"`
}

func (va *AccountValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uva AccountValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	ac := new(types.Account)
	if err := va.unpack(enc, uva.Hint, nil, uva.Balance, uva.Height, uva.ContractAccountStatus); err != nil {
		return err
	} else if err := ac.DecodeJSON(b, enc); err != nil {
		return err
	} else {
		va.ac = *ac

		return nil
	}
}
