package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (fa *FixedFeeer) unpack(enc encoder.Encoder, ht hint.Hint, rc string, am string) error {
	e := util.StringError("failed to unmarshal FixedFeeer")

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fa.receiver = ad
	}

	if big, err := common.NewBigFromString(am); err != nil {
		return e.Wrap(err)
	} else {
		fa.amount = big
	}
	fa.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}

func (fa *RatioFeeer) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	rc string,
	ratio float64,
	min, max string,
) error {
	e := util.StringError("failed to unmarshal RatioFeeer")

	switch ad, err := base.DecodeAddress(rc, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fa.receiver = ad
	}

	fa.ratio = ratio

	if min, err := common.NewBigFromString(min); err != nil {
		return e.Wrap(err)
	} else {
		fa.min = min
	}

	if max, err := common.NewBigFromString(max); err != nil {
		return e.Wrap(err)
	} else {
		fa.max = max
	}
	fa.BaseHinter = hint.NewBaseHinter(ht)

	return nil
}
