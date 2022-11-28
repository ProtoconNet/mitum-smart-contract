package currency

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/util/encoder"
)

func (po *CurrencyPolicy) unpack(enc encoder.Encoder, mn Big, bfe []byte) error {
	po.newAccountMinBalance = mn

	var feeer Feeer
	err := encoder.Decode(enc, bfe, &feeer)
	if err != nil {
		return errors.WithMessage(err, "failed to decode feeer")
	}
	po.feeer = feeer

	return nil
}
