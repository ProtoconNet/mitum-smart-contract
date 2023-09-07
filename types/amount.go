package types

import (
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
)

var (
	AmountHint = hint.MustNewHint("mitum-currency-amount-v0.0.1")
)

type Amount struct {
	hint.BaseHinter
	big common.Big
	cid CurrencyID
}

func NewAmount(big common.Big, cid CurrencyID) Amount {
	am := Amount{BaseHinter: hint.NewBaseHinter(AmountHint), big: big, cid: cid}

	return am
}

func NewZeroAmount(cid CurrencyID) Amount {
	return NewAmount(common.NewBig(0), cid)
}

func MustNewAmount(big common.Big, cid CurrencyID) Amount {
	am := NewAmount(big, cid)
	if err := am.IsValid(nil); err != nil {
		panic(err)
	}

	return am
}

func (am Amount) Bytes() []byte {
	return util.ConcatBytesSlice(
		am.big.Bytes(),
		am.cid.Bytes(),
	)
}

func (am Amount) Hash() util.Hash {
	return am.GenerateHash()
}

func (am Amount) GenerateHash() util.Hash {
	return valuehash.NewSHA256(am.Bytes())
}

func (am Amount) IsEmpty() bool {
	return len(am.cid) < 1 || !am.big.OverNil()
}

func (am Amount) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false,
		am.BaseHinter,
		am.cid,
		am.big,
	); err != nil {
		return util.ErrInvalid.Errorf("validation check of Amount: %v", err)
	}

	return nil
}

func (am Amount) Big() common.Big {
	return am.big
}

func (am Amount) Currency() CurrencyID {
	return am.cid
}

func (am Amount) String() string {
	return fmt.Sprintf("%s(%s)", am.big.String(), am.cid)
}

func (am Amount) Equal(b Amount) bool {
	switch {
	case am.cid != b.cid:
		return false
	case !am.big.Equal(b.big):
		return false
	default:
		return true
	}
}

func (am Amount) WithBig(big common.Big) Amount {
	am.big = big

	return am
}
