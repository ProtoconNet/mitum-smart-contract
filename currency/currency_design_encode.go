package currency

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (de *CurrencyDesign) unpack(enc encoder.Encoder, ht hint.Hint, bam []byte, sga string, bpo []byte, ag string) error {
	e := util.StringErrorFunc("failed to unmarshal CurrencyDesign")

	de.BaseHinter = hint.NewBaseHinter(ht)

	var am Amount
	if err := encoder.Decode(enc, bam, &am); err != nil {
		return errors.WithMessage(err, "failed to decode amount")
	}

	de.amount = am

	switch ad, err := base.DecodeAddress(sga, enc); {
	case err != nil:
		return e(err, "")
	default:
		de.genesisAccount = ad
	}

	var policy CurrencyPolicy

	if err := encoder.Decode(enc, bpo, &policy); err != nil {
		return errors.WithMessage(err, "failed to decode currency policy")
	}

	de.policy = policy

	if big, err := NewBigFromString(ag); err != nil {
		return errors.WithMessage(err, "failed to decode currency policy")
	} else {
		de.aggregate = big
	}

	return nil
}
