package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseWithdrawItem struct {
	hint.BaseHinter
	target  base.Address
	amounts []types.Amount
}

func NewBaseWithdrawItem(ht hint.Hint, target base.Address, amounts []types.Amount) BaseWithdrawItem {
	return BaseWithdrawItem{
		BaseHinter: hint.NewBaseHinter(ht),
		target:     target,
		amounts:    amounts,
	}
}

func (it BaseWithdrawItem) Bytes() []byte {
	bs := make([][]byte, len(it.amounts)+1)
	bs[0] = it.target.Bytes()

	for i := range it.amounts {
		bs[i+1] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseWithdrawItem) IsValid([]byte) error {
	if err := it.BaseHinter.IsValid(nil); err != nil {
		return err
	}

	if err := util.CheckIsValiders(nil, false, it.target); err != nil {
		return err
	}

	if len(it.amounts) < 1 {
		return util.ErrInvalid.Errorf("empty amounts")
	}

	founds := map[types.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return util.ErrInvalid.Errorf("duplicate currency found, %v", am.Currency())
		}
		founds[am.Currency()] = struct{}{}

		if err := am.IsValid(nil); err != nil {
			return err
		} else if !am.Big().OverZero() {
			return util.ErrInvalid.Errorf("amount should be over zero")
		}
	}

	return nil
}

func (it BaseWithdrawItem) Target() base.Address {
	return it.target
}

func (it BaseWithdrawItem) Amounts() []types.Amount {
	return it.amounts
}

func (it BaseWithdrawItem) Rebuild() WithdrawItem {
	ams := make([]types.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
