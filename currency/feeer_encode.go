package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fa *FixedFeeer) unpack(enc encoder.Encoder, src string, am Big) error {
	e := util.StringErrorFunc("failed to unmarshal FixedFeeer")

	switch ad, err := base.DecodeAddress(src, enc); {
	case err != nil:
		return e(err, "")
	default:
		fa.receiver = ad
	}

	fa.amount = am

	return nil
}

func (fa *RatioFeeer) unpack(
	enc encoder.Encoder,
	src string,
	ratio float64,
	min, max Big,
) error {
	e := util.StringErrorFunc("failed to unmarshal RatioFeeer")

	switch ad, err := base.DecodeAddress(src, enc); {
	case err != nil:
		return e(err, "")
	default:
		fa.receiver = ad
	}

	fa.ratio = ratio
	fa.min = min
	fa.max = max

	return nil
}
