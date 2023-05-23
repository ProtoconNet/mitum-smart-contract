package currency

import (
	base2 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *BaseTransfersItem) unpack(enc encoder.Encoder, ht hint.Hint, rc string, bam []byte) error {
	e := util.StringErrorFunc("failed to unmarshal BaseTransfersItem")

	it.BaseHinter = hint.NewBaseHinter(ht)

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e(err, "")
	default:
		it.receiver = ad
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return e(err, "")
	}

	amounts := make([]base2.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(base2.Amount)
		if !ok {
			return util.ErrWrongType.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
