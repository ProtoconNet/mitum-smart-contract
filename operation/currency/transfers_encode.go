package currency

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *TransfersFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringError("failed to unmarshal TransfersFact")

	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.sender = ad
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]TransfersItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(TransfersItem)
		if !ok {
			return errors.Errorf("expected TransfersItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
