package currency

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *SuffrageInflationItem) unpack(enc encoder.Encoder, ht hint.Hint, rc string, bam []byte) error {
	e := util.StringErrorFunc("failed to unmarshal SuffrageInflationItem")

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e(err, "")
	default:
		it.receiver = ad
	}

	if hinter, err := enc.Decode(bam); err != nil {
		return e(err, "")
	} else if am, ok := hinter.(Amount); !ok {
		return util.ErrWrongType.Errorf("expected Amount, not %T", hinter)
	} else {
		it.amount = am
	}
	it.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}
