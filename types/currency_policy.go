package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

var (
	CurrencyPolicyHint = hint.MustNewHint("mitum-currency-currency-policy-v0.0.1")
)

type CurrencyPolicy struct {
	hint.BaseHinter
	minBalance common.Big
	feeer      Feeer
}

func NewCurrencyPolicy(newAccountMinBalance common.Big, feeer Feeer) CurrencyPolicy {
	return CurrencyPolicy{
		BaseHinter: hint.NewBaseHinter(CurrencyPolicyHint),
		minBalance: newAccountMinBalance, feeer: feeer,
	}
}

func (po CurrencyPolicy) Bytes() []byte {
	return util.ConcatBytesSlice(po.minBalance.Bytes(), po.feeer.Bytes())
}

func (po CurrencyPolicy) IsValid([]byte) error {
	if !po.minBalance.OverNil() {
		return common.ErrValueInvalid.Wrap(errors.Errorf("new Account Minimum Balance under zero"))
	}

	if err := util.CheckIsValiders(nil, false, po.BaseHinter, po.feeer); err != nil {
		return common.ErrValueInvalid.Wrap(errors.Errorf("invalid currency policy, %v", err))
	}

	return nil
}

func (po CurrencyPolicy) MinBalance() common.Big {
	return po.minBalance
}

func (po CurrencyPolicy) Feeer() Feeer {
	return po.feeer
}
