package digest

import (
	base3 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (va *AccountValue) unpack(enc encoder.Encoder, ht hint.Hint, bac []byte, bl []byte, height base.Height) error {
	va.BaseHinter = hint.NewBaseHinter(ht)

	ac, err := enc.Decode(bac)
	switch {
	case err != nil:
		return err
	case ac != nil:
		if v, ok := ac.(base3.Account); !ok {
			return util.ErrWrongType.Errorf("expected Account, not %T", ac)
		} else {
			va.ac = v
		}
	}

	hbl, err := enc.DecodeSlice(bl)
	if err != nil {
		return err
	}

	balance := make([]base3.Amount, len(hbl))
	for i := range hbl {
		j, ok := hbl[i].(base3.Amount)
		if !ok {
			return mitumutil.ErrWrongType.Errorf("expected currency.Amount, not %T", hbl[i])
		}
		balance[i] = j
	}

	va.balance = balance
	va.height = height

	return nil
}
