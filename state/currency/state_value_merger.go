package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
	"sync"
)

type BalanceStateValueMerger struct {
	*common.BaseStateValueMerger
	existing BalanceStateValue
	add      common.Big
	remove   common.Big
	currency types.CurrencyID
	sync.Mutex
}

func NewBalanceStateValueMerger(height base.Height, key string, currency types.CurrencyID, st base.State) *BalanceStateValueMerger {
	nst := st
	if st == nil {
		nst = common.NewBaseState(base.NilHeight, key, nil, nil, nil)
	}

	s := &BalanceStateValueMerger{
		BaseStateValueMerger: common.NewBaseStateValueMerger(height, nst.Key(), nst),
	}

	s.existing = NewBalanceStateValue(types.NewZeroAmount(currency))
	if nst.Value() != nil {
		s.existing = nst.Value().(BalanceStateValue) //nolint:forcetypeassert //...
	}
	s.add = common.ZeroBig
	s.remove = common.ZeroBig
	s.currency = currency

	return s
}

func (s *BalanceStateValueMerger) Merge(value base.StateValue, ops util.Hash) error {
	s.Lock()
	defer s.Unlock()

	switch t := value.(type) {
	case AddBalanceStateValue:
		s.add = s.add.Add(t.Amount.Big())
	case DeductBalanceStateValue:
		s.remove = s.remove.Add(t.Amount.Big())
	default:
		return errors.Errorf("unsupported balance state value, %T", value)
	}

	s.AddOperation(ops)

	return nil
}

func (s *BalanceStateValueMerger) CloseValue() (base.State, error) {
	s.Lock()
	defer s.Unlock()

	newValue, err := s.closeValue()
	if err != nil {
		return nil, errors.WithMessage(err, "close BalanceStateValueMerger")
	}

	s.BaseStateValueMerger.SetValue(newValue)

	return s.BaseStateValueMerger.CloseValue()
}

func (s *BalanceStateValueMerger) closeValue() (base.StateValue, error) {
	existingAmount := s.existing.Amount

	if s.add.OverZero() {
		existingAmount = existingAmount.WithBig(existingAmount.Big().Add(s.add))
	}

	if s.remove.OverZero() {
		existingAmount = existingAmount.WithBig(existingAmount.Big().Sub(s.remove))
	}

	return NewBalanceStateValue(
		existingAmount,
	), nil
}
