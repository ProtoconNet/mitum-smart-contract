package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (it *MintItem) unpack(enc encoder.Encoder, ht hint.Hint, rc string, bam []byte) error {
	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return err
	default:
		it.receiver = ad
	}

	if hinter, err := enc.Decode(bam); err != nil {
		return err
	} else if am, ok := hinter.(types.Amount); !ok {
		return common.ErrTypeMismatch.Wrap(errors.Errorf("expected InitialSupply, not %T", hinter))
	} else {
		it.amount = am
	}
	it.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}
