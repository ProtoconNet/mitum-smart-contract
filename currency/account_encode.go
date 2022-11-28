package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/valuehash"
)

func (ac *Account) unpack(enc encoder.Encoder, h valuehash.HashDecoder, sad string, bks []byte) error {
	e := util.StringErrorFunc("failed to unmarshal Account")

	switch ad, err := base.DecodeAddress(sad, enc); {
	case err != nil:
		return e(err, "")
	default:
		ac.address = ad
	}

	k, err := enc.Decode(bks)
	if err != nil {
		return e(err, "")
	} else if k != nil {
		v, ok := k.(BaseAccountKeys)
		if !ok {
			return util.ErrWrongType.Errorf("expected Keys, not %T", k)
		}
		ac.keys = v
	}

	ac.h = h.Hash()

	return nil
}
