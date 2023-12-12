package currency

import (
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
	"strings"
)

var (
	AccountStateValueHint        = hint.MustNewHint("account-state-value-v0.0.1")
	BalanceStateValueHint        = hint.MustNewHint("balance-state-value-v0.0.1")
	CurrencyDesignStateValueHint = hint.MustNewHint("currency-design-state-value-v0.0.1")
)

var (
	StateKeyAccountSuffix        = ":account"
	StateKeyBalanceSuffix        = ":balance"
	StateKeyCurrencyDesignPrefix = "currencydesign:"
)

type AccountStateValue struct {
	hint.BaseHinter
	Account types.Account
}

func NewAccountStateValue(account types.Account) AccountStateValue {
	return AccountStateValue{
		BaseHinter: hint.NewBaseHinter(AccountStateValueHint),
		Account:    account,
	}
}

func (a AccountStateValue) Hint() hint.Hint {
	return a.BaseHinter.Hint()
}

func (a AccountStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid AccountStateValue")

	if err := a.BaseHinter.IsValid(AccountStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, a.Account); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (a AccountStateValue) HashBytes() []byte {
	return a.Account.Bytes()
}

func StateKeysValue(st base.State) (types.AccountKeys, error) {
	ac, err := LoadStateAccountValue(st)
	if err != nil {
		return nil, err
	}
	return ac.Keys(), nil
}

type BalanceStateValue struct {
	hint.BaseHinter
	Amount types.Amount
}

func NewBalanceStateValue(amount types.Amount) BalanceStateValue {
	return BalanceStateValue{
		BaseHinter: hint.NewBaseHinter(BalanceStateValueHint),
		Amount:     amount,
	}
}

func (b BalanceStateValue) Hint() hint.Hint {
	return b.BaseHinter.Hint()
}

func (b BalanceStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid BalanceStateValue")

	if err := b.BaseHinter.IsValid(BalanceStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, b.Amount); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (b BalanceStateValue) HashBytes() []byte {
	return b.Amount.Bytes()
}

func StateBalanceValue(st base.State) (types.Amount, error) {
	v := st.Value()
	if v == nil {
		return types.Amount{}, util.ErrNotFound.Errorf("balance not found in State")
	}

	a, ok := v.(BalanceStateValue)
	if !ok {
		return types.Amount{}, errors.Errorf("invalid balance value found, %T", v)
	}

	return a.Amount, nil
}

type AddBalanceStateValue struct {
	Amount types.Amount
}

func NewAddBalanceStateValue(amount types.Amount) AddBalanceStateValue {
	return AddBalanceStateValue{
		Amount: amount,
	}
}

func (b AddBalanceStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid AddBalanceStateValue")

	if err := util.CheckIsValiders(nil, false, b.Amount); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (b AddBalanceStateValue) HashBytes() []byte {
	return b.Amount.Bytes()
}

type DeductBalanceStateValue struct {
	Amount types.Amount
}

func NewDeductBalanceStateValue(amount types.Amount) DeductBalanceStateValue {
	return DeductBalanceStateValue{
		Amount: amount,
	}
}

func (b DeductBalanceStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid DeductBalanceStateValue")

	if err := util.CheckIsValiders(nil, false, b.Amount); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (b DeductBalanceStateValue) HashBytes() []byte {
	return b.Amount.Bytes()
}

type CurrencyDesignStateValue struct {
	hint.BaseHinter
	CurrencyDesign types.CurrencyDesign
}

func NewCurrencyDesignStateValue(currencyDesign types.CurrencyDesign) CurrencyDesignStateValue {
	return CurrencyDesignStateValue{
		BaseHinter:     hint.NewBaseHinter(CurrencyDesignStateValueHint),
		CurrencyDesign: currencyDesign,
	}
}

func (c CurrencyDesignStateValue) Hint() hint.Hint {
	return c.BaseHinter.Hint()
}

func (c CurrencyDesignStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid CurrencyDesignStateValue")

	if err := c.BaseHinter.IsValid(CurrencyDesignStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, c.CurrencyDesign); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (c CurrencyDesignStateValue) HashBytes() []byte {
	return c.CurrencyDesign.Bytes()
}

func StateCurrencyDesignValue(st base.State) (types.CurrencyDesign, error) {
	v := st.Value()
	if v == nil {
		return types.CurrencyDesign{}, util.ErrNotFound.Errorf("currency design not found in State")
	}

	de, ok := v.(CurrencyDesignStateValue)
	if !ok {
		return types.CurrencyDesign{}, errors.Errorf("invalid currency design value found, %T", v)
	}

	return de.CurrencyDesign, nil
}

func StateBalanceKeyPrefix(a base.Address, cid types.CurrencyID) string {
	return fmt.Sprintf("%s-%s", a.String(), cid)
}

func StateKeyAccount(a base.Address) string {
	return fmt.Sprintf("%s%s", a.String(), StateKeyAccountSuffix)
}

func IsStateAccountKey(key string) bool {
	return strings.HasSuffix(key, StateKeyAccountSuffix)
}

func LoadStateAccountValue(st base.State) (types.Account, error) {
	v := st.Value()
	if v == nil {
		return types.Account{}, util.ErrNotFound.Errorf("account not found in State")
	}

	s, ok := v.(AccountStateValue)
	if !ok {
		return types.Account{}, errors.Errorf("invalid account value found, %T", v)
	}
	return s.Account, nil

}

func StateKeyBalance(a base.Address, cid types.CurrencyID) string {
	return fmt.Sprintf("%s%s", StateBalanceKeyPrefix(a, cid), StateKeyBalanceSuffix)
}

func IsStateBalanceKey(key string) bool {
	return strings.HasSuffix(key, StateKeyBalanceSuffix)
}

func IsStateCurrencyDesignKey(key string) bool {
	return strings.HasPrefix(key, StateKeyCurrencyDesignPrefix)
}

func StateKeyCurrencyDesign(cid types.CurrencyID) string {
	return fmt.Sprintf("%s%s", StateKeyCurrencyDesignPrefix, cid)
}
