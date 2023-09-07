package digest

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (va *AccountValue) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	bac, bl []byte,
	height base.Height,
	cas []byte,
) error {
	va.BaseHinter = hint.NewBaseHinter(ht)
	ac, err := enc.Decode(bac)
	switch {
	case err != nil:
		return err
	case ac != nil:
		if v, ok := ac.(types.Account); !ok {
			return errors.Errorf("expected Account, not %T", ac)
		} else {
			va.ac = v
		}
	}

	hbl, err := enc.DecodeSlice(bl)
	if err != nil {
		return err
	}

	balance := make([]types.Amount, len(hbl))
	for i := range hbl {
		j, ok := hbl[i].(types.Amount)
		if !ok {
			return errors.Errorf("expected currency.Amount, not %T", hbl[i])
		}
		balance[i] = j
	}

	va.balance = balance
	va.height = height

	status, err := enc.Decode(cas)
	switch {
	case err != nil:
		return err
	case status != nil:
		if v, ok := status.(types.ContractAccountStatus); !ok {
			return errors.Errorf("expected ContractAccountStatus, not %T", status)
		} else {
			va.contractAccountStatus = v
		}
	}

	return nil
}
