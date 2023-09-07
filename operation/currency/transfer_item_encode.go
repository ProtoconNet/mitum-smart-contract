package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (it *BaseTransferItem) unpack(enc encoder.Encoder, ht hint.Hint, rc string, bam []byte) error {
	e := util.StringError("unmarshal BaseTransferItem")

	it.BaseHinter = hint.NewBaseHinter(ht)

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		it.receiver = ad
	}

	ham, err := enc.DecodeSlice(bam)
	if err != nil {
		return e.Wrap(err)
	}

	amounts := make([]types.Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(types.Amount)
		if !ok {
			return errors.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
