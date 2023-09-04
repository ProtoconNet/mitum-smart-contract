package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *RegisterCurrencyFact) unpack(
	enc encoder.Encoder,
	bcr []byte,
) error {
	e := util.StringError("failed to unmarshal RegisterCurrencyFact")

	if hinter, err := enc.Decode(bcr); err != nil {
		return e.Wrap(err)
	} else if cr, ok := hinter.(types.CurrencyDesign); !ok {
		return errors.Errorf("expected CurrencyDesign not %T,", hinter)
	} else {
		fact.currency = cr
	}

	return nil
}
