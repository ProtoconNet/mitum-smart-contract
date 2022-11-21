package digest

import (
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	mitumutil "github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (va *AccountValue) unpack(enc encoder.Encoder, bac []byte, bl []byte, height base.Height) error {
	ac, err := enc.Decode(bac)
	switch {
	case err != nil:
		return err
	case ac != nil:
		if v, ok := ac.(currency.Account); !ok {
			return util.ErrWrongType.Errorf("expected Account, not %T", ac)
		} else {
			va.ac = v
		}
	}

	hbl, err := enc.DecodeSlice(bl)
	if err != nil {
		return err
	}

	balance := make([]currency.Amount, len(hbl))
	for i := range hbl {
		j, ok := hbl[i].(currency.Amount)
		if !ok {
			return mitumutil.ErrWrongType.Errorf("expected currency.Amount, not %T", hbl[i])
		}
		balance[i] = j
	}

	va.balance = balance
	va.height = height

	return nil
}
