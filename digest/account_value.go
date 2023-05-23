package digest

import (
	base3 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	AccountValueHint = hint.MustNewHint("mitum-currency-account-value-v0.0.1")
)

type AccountValue struct {
	hint.BaseHinter
	ac      base3.Account
	balance []base3.Amount
	height  base.Height
}

func NewAccountValue(st base.State) (AccountValue, error) {
	var ac base3.Account
	switch a, ok, err := IsAccountState(st); {
	case err != nil:
		return AccountValue{}, err
	case !ok:
		return AccountValue{}, errors.Errorf("not state for currency.Account, %T", st.Value())
	default:
		ac = a
	}

	return AccountValue{
		BaseHinter: hint.NewBaseHinter(AccountValueHint),
		ac:         ac,
		height:     st.Height(),
	}, nil
}

func (va AccountValue) Account() base3.Account {
	return va.ac
}

func (va AccountValue) Balance() []base3.Amount {
	return va.balance
}

func (va AccountValue) Height() base.Height {
	return va.height
}

func (va AccountValue) SetHeight(height base.Height) AccountValue {
	va.height = height

	return va
}

func (va AccountValue) SetBalance(balance []base3.Amount) AccountValue {
	va.balance = balance

	return va
}
