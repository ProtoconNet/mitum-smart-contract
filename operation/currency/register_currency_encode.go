package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *RegisterCurrencyFact) unpack(
	enc encoder.Encoder,
	bcr []byte,
) error {
	if hinter, err := enc.Decode(bcr); err != nil {
		return err
	} else if cr, ok := hinter.(types.CurrencyDesign); !ok {
		return common.ErrTypeMismatch.Wrap(errors.Errorf("expected CurrencyDesign not %T,", hinter))
	} else {
		fact.currency = cr
	}

	return nil
}
