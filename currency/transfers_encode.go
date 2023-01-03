package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *TransfersFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringErrorFunc("failed to unmarshal TransfersFact")

	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = ad
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e(err, "")
	}

	items := make([]TransfersItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(TransfersItem)
		if !ok {
			return util.ErrWrongType.Errorf("expected TransfersItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
