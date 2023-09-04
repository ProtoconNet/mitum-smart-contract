package currency

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *TransferFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringError("failed to unmarshal TransferFact")

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

	items := make([]TransferItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(TransferItem)
		if !ok {
			return errors.Errorf("expected TransferItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
