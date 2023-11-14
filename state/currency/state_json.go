package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type AccountStateValueJSONMarshaler struct {
	hint.BaseHinter
	Account types.Account `json:"account"`
}

func (a AccountStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AccountStateValueJSONMarshaler{
		BaseHinter: a.BaseHinter,
		Account:    a.Account,
	})
}

type AccountStateValueJSONUnmarshaler struct {
	AC json.RawMessage `json:"account"`
}

func (a *AccountStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode AccountStateValue")

	var u AccountStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	var ac types.Account

	if err := ac.DecodeJSON(u.AC, enc); err != nil {
		return e.Wrap(err)
	}

	a.Account = ac

	return nil
}

type BalanceStateValueJSONMarshaler struct {
	hint.BaseHinter
	Amount types.Amount `json:"amount"`
}

func (b BalanceStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BalanceStateValueJSONMarshaler{
		BaseHinter: b.BaseHinter,
		Amount:     b.Amount,
	})
}

type BalanceStateValueJSONUnmarshaler struct {
	AM json.RawMessage `json:"amount"`
}

func (b *BalanceStateValue) DecodeJSON(v []byte, enc encoder.Encoder) error {
	e := util.StringError("decode BalanceStateValue")

	var u BalanceStateValueJSONUnmarshaler
	if err := enc.Unmarshal(v, &u); err != nil {
		return e.Wrap(err)
	}

	var am types.Amount

	if err := am.DecodeJSON(u.AM, enc); err != nil {
		return e.Wrap(err)
	}

	b.Amount = am

	return nil
}

type CurrencyDesignStateValueJSONMarshaler struct {
	hint.BaseHinter
	CurrencyDesign types.CurrencyDesign `json:"currencydesign"`
}

func (c CurrencyDesignStateValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyDesignStateValueJSONMarshaler{
		BaseHinter:     c.BaseHinter,
		CurrencyDesign: c.CurrencyDesign,
	})
}

type CurrencyDesignStateValueJSONUnmarshaler struct {
	CD json.RawMessage `json:"currencydesign"`
}

func (c *CurrencyDesignStateValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode CurrencyDesignStateValue")

	var u CurrencyDesignStateValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &u); err != nil {
		return e.Wrap(err)
	}

	var cd types.CurrencyDesign

	if err := cd.DecodeJSON(u.CD, enc); err != nil {
		return e.Wrap(err)
	}

	c.CurrencyDesign = cd

	return nil
}
