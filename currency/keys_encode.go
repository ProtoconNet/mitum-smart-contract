package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (ky *BaseAccountKey) unpack(enc encoder.Encoder, ht hint.Hint, w uint, kd string) error {
	e := util.StringErrorFunc("failed to unmarshal BaseAccountKey")

	ky.BaseHinter = hint.NewBaseHinter(ht)
	switch pk, err := base.DecodePublickeyFromString(kd, enc); {
	case err != nil:
		return e(err, "")
	default:
		ky.k = pk
	}
	ky.w = w

	return nil
}
