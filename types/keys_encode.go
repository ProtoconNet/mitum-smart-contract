package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (ky *BaseAccountKey) unpack(enc encoder.Encoder, ht hint.Hint, w uint, sk string) error {
	ky.BaseHinter = hint.NewBaseHinter(ht)
	switch pk, err := base.DecodePublickeyFromString(sk, enc); {
	case err != nil:
		return err
	default:
		ky.k = pk
	}
	ky.w = w

	return nil
}

func (ks *BaseAccountKeys) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	h util.Hash,
	bks []byte,
	th uint,
) error {
	ks.BaseHinter = hint.NewBaseHinter(ht)

	hks, err := enc.DecodeSlice(bks)
	if err != nil {
		return err
	}

	keys := make([]AccountKey, len(hks))
	for i := range hks {
		j, ok := hks[i].(BaseAccountKey)
		if !ok {
			return errors.Errorf("expected BaseAccountKey, not %T", hks[i])
		}

		keys[i] = j
	}
	ks.keys = keys

	ks.h = h

	ks.threshold = th

	return nil
}

func (ks *NilAccountKeys) unpack(
	_ encoder.Encoder,
	ht hint.Hint,
	h util.Hash,
	th uint,
) error {
	ks.BaseHinter = hint.NewBaseHinter(ht)
	ks.h = h
	ks.threshold = th

	return nil
}

func (ks *ContractAccountKeys) unpack(enc encoder.Encoder, ht hint.Hint, h common.HashDecoder, bks []byte, th uint) error {
	ks.BaseHinter = hint.NewBaseHinter(ht)

	hks, err := enc.DecodeSlice(bks)
	if err != nil {
		return err
	}

	keys := make([]AccountKey, len(hks))
	for i := range hks {
		j, ok := hks[i].(BaseAccountKey)
		if !ok {
			return errors.Errorf("expected BaseAccountKey, not %T", hks[i])
		}

		keys[i] = j
	}
	ks.keys = keys

	ks.h = h.Hash()
	ks.threshold = th

	return nil
}
