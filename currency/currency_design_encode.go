package currency

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (de *CurrencyDesign) unpack(enc encoder.Encoder, bam []byte, sga string, bpo []byte, ag Big) error {
	e := util.StringErrorFunc("failed to unmarshal CurrencyDesign")

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
	de.aggregate = ag

	return nil
}
