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
	status types.ContractAccountStatus
}

func NewContractAccountStateValue(status types.ContractAccountStatus) ContractAccountStateValue {
	return ContractAccountStateValue{
		BaseHinter: hint.NewBaseHinter(ContractAccountStateValueHint),
		status:     status,
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

	if err := util.CheckIsValiders(nil, false, c.status); err != nil {
		return e.Wrap(err)
	}

	return nil
}

func (c ContractAccountStateValue) HashBytes() []byte {
	return c.status.Bytes()
}

func (c ContractAccountStateValue) Status() types.ContractAccountStatus {
	return c.status
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
	return cs.status, nil
}
