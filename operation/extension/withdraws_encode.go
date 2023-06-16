package extension

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *WithdrawsFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringError("failed to unmarshal WithdrawsFact")

	switch a, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]WithdrawsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(WithdrawsItem)
		if !ok {
			return e.Wrap(errors.Errorf("expected WithdrawsItem, not %T", hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
