package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (it *MintItem) unpack(enc encoder.Encoder, ht hint.Hint, rc string, bam []byte) error {
	e := util.StringError("failed to unmarshal MintItem")

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.receiver = ad
	}

	if hinter, err := enc.Decode(bam); err != nil {
		return e.Wrap(err)
	} else if am, ok := hinter.(types.Amount); !ok {
		return errors.Errorf("expected Amount, not %T", hinter)
	} else {
		it.amount = am
	}
	it.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}
