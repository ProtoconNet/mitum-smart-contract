package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseTransferItem struct {
	hint.BaseHinter
	receiver base.Address
	amounts  []types.Amount
}

func NewBaseTransferItem(ht hint.Hint, receiver base.Address, amounts []types.Amount) BaseTransferItem {
	return BaseTransferItem{
		BaseHinter: hint.NewBaseHinter(ht),
		receiver:   receiver,
		amounts:    amounts,
	}
}

func (it BaseTransferItem) Bytes() []byte {
	bs := make([][]byte, len(it.amounts)+1)
	bs[0] = it.receiver.Bytes()

	for i := range it.amounts {
		bs[i+1] = it.amounts[i].Bytes()
	}

	return util.ConcatBytesSlice(bs...)
}

func (it BaseTransferItem) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, it.receiver); err != nil {
		return err
	}

	if n := len(it.amounts); n == 0 {
		return util.ErrInvalid.Errorf("empty amounts")
	}

	founds := map[types.CurrencyID]struct{}{}
	for i := range it.amounts {
		am := it.amounts[i]
		if _, found := founds[am.Currency()]; found {
			return util.ErrInvalid.Errorf("duplicated currency found, %v", am.Currency())
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

func (it BaseTransferItem) Receiver() base.Address {
	return it.receiver
}

func (it BaseTransferItem) Amounts() []types.Amount {
	return it.amounts
}

func (it BaseTransferItem) Rebuild() TransferItem {
	ams := make([]types.Amount, len(it.amounts))
	for i := range it.amounts {
		am := it.amounts[i]
		ams[i] = am.WithBig(am.Big())
	}

	it.amounts = ams

	return it
}
