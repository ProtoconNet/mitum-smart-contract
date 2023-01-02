package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
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

func (ks *BaseAccountKeys) unpack(enc encoder.Encoder, ht hint.Hint, h valuehash.HashDecoder, bks []byte, th uint) error {
	e := util.StringErrorFunc("failed to unmarshal BaseAccountKeys")

	ks.BaseHinter = hint.NewBaseHinter(ht)

	hks, err := enc.DecodeSlice(bks)
	if err != nil {
		return e(err, "")
	}

	keys := make([]AccountKey, len(hks))
	for i := range hks {
		j, ok := hks[i].(BaseAccountKey)
		if !ok {
			return util.ErrWrongType.Errorf("expected BaseAccountKey, not %T", hks[i])
		}

		keys[i] = j
	}
	ks.keys = keys

	ks.h = h.Hash()
	ks.threshold = th

	return nil
}
