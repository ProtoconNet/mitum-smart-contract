package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/valuehash"
	"github.com/pkg/errors"
)

func (ac *Account) unpack(enc encoder.Encoder, ht hint.Hint, h valuehash.HashDecoder, ad string, bks []byte) error {
	ac.BaseHinter = hint.NewBaseHinter(ht)
	switch ad, err := base.DecodeAddress(ad, enc); {
	case err != nil:
		return err
	default:
		ac.address = ad
	}

	k, err := enc.Decode(bks)
	if err != nil {
		return err
	} else if k != nil {
		v, ok := k.(AccountKeys)
		if !ok {
			return errors.Errorf("expected AccountKeys, not %T", k)
		}
		ac.keys = v
	}

	ac.h = h.Hash()

	return nil
}
