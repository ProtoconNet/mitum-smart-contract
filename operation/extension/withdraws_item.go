package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseWithdrawsItem struct {
	hint.BaseHinter
	target  mitumbase.Address
	amounts []base.Amount
}

func NewBaseWithdrawsItem(ht hint.Hint, target mitumbase.Address, amounts []base.Amount) BaseWithdrawsItem {
	return BaseWithdrawsItem{
		BaseHinter: hint.NewBaseHinter(ht),
		target:     target,
		amounts:    amounts,
	}
}

func (it BaseWithdrawsItem) Bytes() []byte {
	bs := make([][]byte, len(it.amounts)+1)
	bs[0] = it.target.Bytes()

	for i := range it.amounts {
		bs[i+1] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseWithdrawsItem) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, it.target); err != nil {
		return err
	}

	if n := len(it.amounts); n == 0 {
		return util.ErrInvalid.Errorf("empty amounts")
	}

	founds := map[base.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return util.ErrInvalid.Errorf("duplicate currency found, %q", am.Currency())
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

func (it BaseWithdrawsItem) Target() mitumbase.Address {
	return it.target
}

func (it BaseWithdrawsItem) Amounts() []base.Amount {
	return it.amounts
}

func (it BaseWithdrawsItem) Rebuild() WithdrawsItem {
	ams := make([]base.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
