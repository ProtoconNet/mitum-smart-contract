package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (it *BaseWithdrawItem) unpack(enc encoder.Encoder, ht hint.Hint, tg string, bam []byte) error {
	it.BaseHinter = hint.NewBaseHinter(ht)

	switch a, err := base.DecodeAddress(tg, enc); {
	case err != nil:
		return err
	default:
		it.target = a
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return err
	}

	amounts := make([]types.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(types.Amount)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected Amount, not %T", ham[i]))
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
