package currency

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/hint"
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
	Account Account
}

func NewAccountStateValue(account Account) AccountStateValue {
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

func StateKeysValue(st base.State) (AccountKeys, error) {
	ac, err := LoadStateAccountValue(st)
	if err != nil {
		return nil, err
	}
	return ac.Keys(), nil
}

type BalanceStateValue struct {
	hint.BaseHinter
	Amount Amount
}

func NewBalanceStateValue(amount Amount) BalanceStateValue {
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

func StateBalanceValue(st base.State) (Amount, error) {
	v := st.Value()
	if v == nil {
		return Amount{}, util.ErrNotFound.Errorf("balance not found in State")
	}

	a, ok := v.(BalanceStateValue)
	if !ok {
		return Amount{}, errors.Errorf("invalid balance value found, %T", v)
	}

	return a.Amount, nil
}

type CurrencyDesignStateValue struct {
	hint.BaseHinter
	CurrencyDesign CurrencyDesign
}

func NewCurrencyDesignStateValue(currencyDesign CurrencyDesign) CurrencyDesignStateValue {
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

func StateCurrencyDesignValue(st base.State) (CurrencyDesign, error) {
	v := st.Value()
	if v == nil {
		return CurrencyDesign{}, util.ErrNotFound.Errorf("currency design not found in State")
	}

	de, ok := v.(CurrencyDesignStateValue)
	if !ok {
		return CurrencyDesign{}, errors.Errorf("invalid currency design value found, %T", v)
	}

	return de.CurrencyDesign, nil
}

func StateBalanceKeyPrefix(a base.Address, cid CurrencyID) string {
	return fmt.Sprintf("%s-%s", a.String(), cid)
}

func StateKeyAccount(a base.Address) string {
	return fmt.Sprintf("%s%s", a.String(), StateKeyAccountSuffix)
}

func IsStateAccountKey(key string) bool {
	return strings.HasSuffix(key, StateKeyAccountSuffix)
}

func LoadStateAccountValue(st base.State) (Account, error) {
	v := st.Value()
	if v == nil {
		return Account{}, util.ErrNotFound.Errorf("account not found in State")
	}

	s, ok := v.(AccountStateValue)
	if !ok {
		return Account{}, errors.Errorf("invalid account value found, %T", v)
	}
	return s.Account, nil

}

func StateKeyBalance(a base.Address, cid CurrencyID) string {
	return fmt.Sprintf("%s%s", StateBalanceKeyPrefix(a, cid), StateKeyBalanceSuffix)
}

func IsStateBalanceKey(key string) bool {
	return strings.HasSuffix(key, StateKeyBalanceSuffix)
}

func IsStateCurrencyDesignKey(key string) bool {
	return strings.HasPrefix(key, StateKeyCurrencyDesignPrefix)
}

func StateKeyCurrencyDesign(cid CurrencyID) string {
	return fmt.Sprintf("%s%s", StateKeyCurrencyDesignPrefix, cid)
}

func checkExistsState(
	key string,
	getState base.GetStateFunc,
) error {
	switch _, found, err := getState(key); {
	case err != nil:
		return err
	case !found:
		return base.NewBaseOperationProcessReasonError("state, %q does not exist", key)
	default:
		return nil
	}
}

func existsState(
	k,
	name string,
	getState base.GetStateFunc,
) (base.State, error) {
	switch st, found, err := getState(k); {
	case err != nil:
		return nil, err
	case !found:
		return nil, base.NewBaseOperationProcessReasonError("%s does not exist", name)
	default:
		return st, nil
	}
}

func notExistsState(
	k,
	name string,
	getState base.GetStateFunc,
) (base.State, error) {
	var st base.State
	switch _, found, err := getState(k); {
	case err != nil:
		return nil, err
	case found:
		return nil, base.NewBaseOperationProcessReasonError("%s already exists", name)
	case !found:
		st = NewBaseState(base.NilHeight, k, nil, nil, nil)
	}
	return st, nil
}

func existsCurrencyPolicy(cid CurrencyID, getStateFunc base.GetStateFunc) (CurrencyPolicy, error) {
	var policy CurrencyPolicy
	switch i, found, err := getStateFunc(StateKeyCurrencyDesign(cid)); {
	case err != nil:
		return CurrencyPolicy{}, err
	case !found:
		return CurrencyPolicy{}, base.NewBaseOperationProcessReasonError("currency not found, %v", cid)
	default:
		currencydesign, ok := i.Value().(CurrencyDesignStateValue) //nolint:forcetypeassert //...
		if !ok {
			return CurrencyPolicy{}, errors.Errorf("expected CurrencyDesignStateValue, not %T", i.Value())
		}
		policy = currencydesign.CurrencyDesign.policy
	}
	return policy, nil
}

type AccountStateValueMerger struct {
	*BaseStateValueMerger
}

func NewAccountStateValueMerger(height base.Height, key string, st base.State) *AccountStateValueMerger {
	s := &AccountStateValueMerger{
		BaseStateValueMerger: NewBaseStateValueMerger(height, key, st),
	}

	return s
}

type BalanceStateValueMerger struct {
	*BaseStateValueMerger
}

func NewBalanceStateValueMerger(height base.Height, key string, st base.State) *BalanceStateValueMerger {
	s := &BalanceStateValueMerger{
		BaseStateValueMerger: NewBaseStateValueMerger(height, key, st),
	}

	return s
}

type CurrencyDesignStateValueMerger struct {
	*BaseStateValueMerger
}

func NewCurrencyDesignStateValueMerger(height base.Height, key string, st base.State) *CurrencyDesignStateValueMerger {
	s := &CurrencyDesignStateValueMerger{
		BaseStateValueMerger: NewBaseStateValueMerger(height, key, st),
	}

	return s
}

func NewBalanceStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
	return NewBaseStateMergeValue(
		key,
		stv,
		func(height base.Height, st base.State) base.StateValueMerger {
			return NewBalanceStateValueMerger(height, key, st)
		},
	)
}

func NewAccountStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
	return NewBaseStateMergeValue(
		key,
		stv,
		func(height base.Height, st base.State) base.StateValueMerger {
			return NewAccountStateValueMerger(height, key, st)
		},
	)
}

func NewCurrencyDesignStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
	return NewBaseStateMergeValue(
		key,
		stv,
		func(height base.Height, st base.State) base.StateValueMerger {
			return NewCurrencyDesignStateValueMerger(height, key, st)
		},
	)
}
