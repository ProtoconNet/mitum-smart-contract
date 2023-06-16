package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type AccountStateValueJSONMarshaler struct {
	hint.BaseHinter
	Account types.Account `json:"account"`
}

func (s AccountStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AccountStateValueJSONMarshaler{
		BaseHinter: s.BaseHinter,
		Account:    s.Account,
	})
}

type AccountStateValueJSONUnmarshaler struct {
	AC json.RawMessage `json:"account"`
}

func (s *AccountStateValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode AccountStateValue")

	var u AccountStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	var ac types.Account

	if err := ac.DecodeJSON(u.AC, enc); err != nil {
		return e.Wrap(err)
	}

	s.Account = ac

	return nil
}

type BalanceStateValueJSONMarshaler struct {
	hint.BaseHinter
	Amount types.Amount `json:"amount"`
}

func (s BalanceStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BalanceStateValueJSONMarshaler{
		BaseHinter: s.BaseHinter,
		Amount:     s.Amount,
	})
}

type BalanceStateValueJSONUnmarshaler struct {
	AM json.RawMessage `json:"amount"`
}

func (s *BalanceStateValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode BalanceStateValue")

	var u BalanceStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	var am types.Amount

	if err := am.DecodeJSON(u.AM, enc); err != nil {
		return e.Wrap(err)
	}

	s.Amount = am

	return nil
}

type CurrencyDesignStateValueJSONMarshaler struct {
	hint.BaseHinter
	CurrencyDesign types.CurrencyDesign `json:"currencydesign"`
}

func (s CurrencyDesignStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyDesignStateValueJSONMarshaler{
		BaseHinter:     s.BaseHinter,
		CurrencyDesign: s.CurrencyDesign,
	})
}

type CurrencyDesignStateValueJSONUnmarshaler struct {
	CD json.RawMessage `json:"currencydesign"`
}

func (s *CurrencyDesignStateValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode CurrencyDesignStateValue")

	var u CurrencyDesignStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	var cd types.CurrencyDesign

	if err := cd.DecodeJSON(u.CD, enc); err != nil {
		return e.Wrap(err)
	}

	s.CurrencyDesign = cd

	return nil
}
