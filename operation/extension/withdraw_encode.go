package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *WithdrawFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	switch a, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return err
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return err
	}

	items := make([]WithdrawItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(WithdrawItem)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected WithdrawItem, not %T", hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
