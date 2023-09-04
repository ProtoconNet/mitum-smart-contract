package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	MintItemHint = hint.MustNewHint("mitum-currency-mint-item-v0.0.1")
)

type MintItem struct {
	hint.BaseHinter
	receiver base.Address
	amount   types.Amount
}

func NewMintItem(receiver base.Address, amount types.Amount) MintItem {
	return MintItem{
		BaseHinter: hint.NewBaseHinter(MintItemHint),
		receiver:   receiver,
		amount:     amount,
	}
}

func (it MintItem) Bytes() []byte {
	var br []byte
	if it.receiver != nil {
		br = it.receiver.Bytes()
	}

	return util.ConcatBytesSlice(br, it.amount.Bytes())
}

func (it MintItem) Receiver() base.Address {
	return it.receiver
}

func (it MintItem) Currency() types.CurrencyID {
	return it.amount.Currency()
}

func (it MintItem) Amount() types.Amount {
	return it.amount
}

func (it MintItem) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, it.receiver, it.amount); err != nil {
		return err
	}

	if !it.amount.Big().OverZero() {
		return util.ErrInvalid.Errorf("under zero amount of MintItem")
	}

	return nil
}
