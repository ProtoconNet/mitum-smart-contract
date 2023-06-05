package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/base"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (it *BaseWithdrawsItem) unpack(enc encoder.Encoder, ht hint.Hint, tg string, bam []byte) error {
	e := util.StringErrorFunc("failed to unmarshal BaseWithdrawsItem")

	it.BaseHinter = hint.NewBaseHinter(ht)

	switch a, err := mitumbase.DecodeAddress(tg, enc); {
	case err != nil:
		return e(err, "")
	default:
		it.target = a
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return e(err, "")
	}

	amounts := make([]base.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(base.Amount)
		if !ok {
			return e(util.ErrWrongType.Errorf("expected Amount, not %T", ham[i]), "")
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
