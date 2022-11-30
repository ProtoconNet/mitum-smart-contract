package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (ky *BaseAccountKey) unpack(enc encoder.Encoder, w uint, kd string) error {
	e := util.StringErrorFunc("failed to unmarshal BaseAccountKey")

	switch pk, err := base.DecodePublickeyFromString(kd, enc); {
	case err != nil:
		return e(err, "")
	default:
		ky.k = pk
	}
	ky.w = w

	return nil
}

func (ks *BaseAccountKeys) unpack(enc encoder.Encoder /*h valuehash.HashDecoder, */, bks []byte, th uint) error {
	hks, err := enc.DecodeSlice(bks)
	if err != nil {
		return err
	}

	keys := make([]AccountKey, len(hks))
	for i := range hks {
		j, ok := hks[i].(BaseAccountKey)
		if !ok {
			return util.ErrWrongType.Errorf("expected Key, not %T", hks[i])
		}

		keys[i] = j
	}
	ks.keys = keys

	// ks.h = h.Hash()
	ks.threshold = th

	return nil
}
