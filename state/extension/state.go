package extension

import (
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
	"strings"
)

var ContractAccountStateValueHint = hint.MustNewHint("contract-account-state-value-v0.0.1")

var StateKeyContractAccountSuffix = ":contractaccount"

type ContractAccountStateValue struct {
	hint.BaseHinter
	account types.ContractAccountStatus
}

func NewContractAccountStateValue(account types.ContractAccountStatus) ContractAccountStateValue {
	return ContractAccountStateValue{
		BaseHinter: hint.NewBaseHinter(ContractAccountStateValueHint),
		account:    account,
	}
}

func (c ContractAccountStateValue) Hint() hint.Hint {
	return c.BaseHinter.Hint()
}

func (c ContractAccountStateValue) IsValid([]byte) error {
	e := util.ErrInvalid.Errorf("invalid ContractAccountStateValue")

	if err := c.BaseHinter.IsValid(ContractAccountStateValueHint.Type().Bytes()); err != nil {
		return e.Wrap(err)
	}

	if err := util.CheckIsValiders(nil, false, c.account); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (c ContractAccountStateValue) HashBytes() []byte {
	return c.account.Bytes()
}

func StateKeyContractAccount(a base.Address) string {
	return fmt.Sprintf("%s%s", a.String(), StateKeyContractAccountSuffix)
}

func IsStateContractAccountKey(key string) bool {
	return strings.HasSuffix(key, StateKeyContractAccountSuffix)
}

func StateContractAccountValue(st base.State) (types.ContractAccountStatus, error) {
	v := st.Value()
	if v == nil {
		return types.ContractAccountStatus{}, util.ErrNotFound.Errorf("contract account status not found in State")
	}

	cs, ok := v.(ContractAccountStateValue)
	if !ok {
		return types.ContractAccountStatus{}, errors.Errorf("invalid contract account status value found, %T", v)
	}
	return cs.account, nil
}

//
//type CurrencyDesignStateValueMerger struct {
//	*base.BaseStateValueMerger
//}
//
//func NewCurrencyDesignStateValueMerger(height base.Height, key string, st base.State) *CurrencyDesignStateValueMerger {
//	s := &CurrencyDesignStateValueMerger{
//		BaseStateValueMerger: base.NewBaseStateValueMerger(height, key, st),
//	}
//
//	return s
//}
//
//func NewCurrencyDesignStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
//	return base.NewBaseStateMergeValue(
//		key,
//		stv,
//		func(height base.Height, st base.State) base.StateValueMerger {
//			return NewCurrencyDesignStateValueMerger(height, key, st)
//		},
//	)
//}
//
//type ContractAccountStateValueMerger struct {
//	*base.BaseStateValueMerger
//}
//
//func NewContractAccountStateValueMerger(height base.Height, key string, st base.State) *ContractAccountStateValueMerger {
//	s := &ContractAccountStateValueMerger{
//		BaseStateValueMerger: base.NewBaseStateValueMerger(height, key, st),
//	}
//
//	return s
//}
//
//func NewContractAccountStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
//	return base.NewBaseStateMergeValue(
//		key,
//		stv,
//		func(height base.Height, st base.State) base.StateValueMerger {
//			return NewContractAccountStateValueMerger(height, key, st)
//		},
//	)
//}
