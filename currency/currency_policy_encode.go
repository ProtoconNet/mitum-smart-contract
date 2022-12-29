package currency

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/util/encoder"
	"github.com/spikeekips/mitum/util/hint"
)

func (po *CurrencyPolicy) unpack(enc encoder.Encoder, ht hint.Hint, mn string, bfe []byte) error {
	if big, err := NewBigFromString(mn); err != nil {
		return errors.WithMessage(err, "failed to decode feeer")
	} else {
		po.newAccountMinBalance = big
	}

	po.BaseHinter = hint.NewBaseHinter(ht)
	var feeer Feeer
	err := encoder.Decode(enc, bfe, &feeer)
	if err != nil {
		return errors.WithMessage(err, "failed to decode feeer")
	}
	po.feeer = feeer

	return nil
}
