package currency

import (
	"fmt"
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"strings"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
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
	Account base2.Account
}

func NewAccountStateValue(account base2.Account) AccountStateValue {
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

func StateKeysValue(st base.State) (base2.AccountKeys, error) {
	ac, err := LoadStateAccountValue(st)
	if err != nil {
		return nil, err
	}
	return ac.Keys(), nil
}

type BalanceStateValue struct {
	hint.BaseHinter
	Amount base2.Amount
}

func NewBalanceStateValue(amount base2.Amount) BalanceStateValue {
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

func StateBalanceValue(st base.State) (base2.Amount, error) {
	v := st.Value()
	if v == nil {
		return base2.Amount{}, util.ErrNotFound.Errorf("balance not found in State")
	}

	a, ok := v.(BalanceStateValue)
	if !ok {
		return base2.Amount{}, errors.Errorf("invalid balance value found, %T", v)
	}

	return a.Amount, nil
}

type CurrencyDesignStateValue struct {
	hint.BaseHinter
	CurrencyDesign base2.CurrencyDesign
}

func NewCurrencyDesignStateValue(currencyDesign base2.CurrencyDesign) CurrencyDesignStateValue {
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

func StateCurrencyDesignValue(st base.State) (base2.CurrencyDesign, error) {
	v := st.Value()
	if v == nil {
		return base2.CurrencyDesign{}, util.ErrNotFound.Errorf("currency design not found in State")
	}

	de, ok := v.(CurrencyDesignStateValue)
	if !ok {
		return base2.CurrencyDesign{}, errors.Errorf("invalid currency design value found, %T", v)
	}

	return de.CurrencyDesign, nil
}

func StateBalanceKeyPrefix(a base.Address, cid base2.CurrencyID) string {
	return fmt.Sprintf("%s-%s", a.String(), cid)
}

func StateKeyAccount(a base.Address) string {
	return fmt.Sprintf("%s%s", a.String(), StateKeyAccountSuffix)
}

func IsStateAccountKey(key string) bool {
	return strings.HasSuffix(key, StateKeyAccountSuffix)
}

func LoadStateAccountValue(st base.State) (base2.Account, error) {
	v := st.Value()
	if v == nil {
		return base2.Account{}, util.ErrNotFound.Errorf("account not found in State")
	}

	s, ok := v.(AccountStateValue)
	if !ok {
		return base2.Account{}, errors.Errorf("invalid account value found, %T", v)
	}
	return s.Account, nil

}

func StateKeyBalance(a base.Address, cid base2.CurrencyID) string {
	return fmt.Sprintf("%s%s", StateBalanceKeyPrefix(a, cid), StateKeyBalanceSuffix)
}

func IsStateBalanceKey(key string) bool {
	return strings.HasSuffix(key, StateKeyBalanceSuffix)
}

func IsStateCurrencyDesignKey(key string) bool {
	return strings.HasPrefix(key, StateKeyCurrencyDesignPrefix)
}

func StateKeyCurrencyDesign(cid base2.CurrencyID) string {
	return fmt.Sprintf("%s%s", StateKeyCurrencyDesignPrefix, cid)
}

//
//type AccountStateValueMerger struct {
//	*base2.BaseStateValueMerger
//}
//
//func NewAccountStateValueMerger(height base.Height, key string, st base.State) *AccountStateValueMerger {
//	s := &AccountStateValueMerger{
//		BaseStateValueMerger: base2.NewBaseStateValueMerger(height, key, st),
//	}
//
//	return s
//}
//
//type BalanceStateValueMerger struct {
//	*base2.BaseStateValueMerger
//}
//
//func NewBalanceStateValueMerger(height base.Height, key string, st base.State) *BalanceStateValueMerger {
//	s := &BalanceStateValueMerger{
//		BaseStateValueMerger: base2.NewBaseStateValueMerger(height, key, st),
//	}
//
//	return s
//}
//
//type CurrencyDesignStateValueMerger struct {
//	*base2.BaseStateValueMerger
//}
//
//func NewCurrencyDesignStateValueMerger(height base.Height, key string, st base.State) *CurrencyDesignStateValueMerger {
//	s := &CurrencyDesignStateValueMerger{
//		BaseStateValueMerger: base2.NewBaseStateValueMerger(height, key, st),
//	}
//
//	return s
//}
//
//func NewBalanceStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
//	return base2.NewBaseStateMergeValue(
//		key,
//		stv,
//		func(height base.Height, st base.State) base.StateValueMerger {
//			return NewBalanceStateValueMerger(height, key, st)
//		},
//	)
//}
//
//func NewAccountStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
//	return base2.NewBaseStateMergeValue(
//		key,
//		stv,
//		func(height base.Height, st base.State) base.StateValueMerger {
//			return NewAccountStateValueMerger(height, key, st)
//		},
//	)
//}
//
//func NewCurrencyDesignStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
//	return base2.NewBaseStateMergeValue(
//		key,
//		stv,
//		func(height base.Height, st base.State) base.StateValueMerger {
//			return NewCurrencyDesignStateValueMerger(height, key, st)
//		},
//	)
//}
