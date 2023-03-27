package currency

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CurrencyRegisterFact) unpack(
	enc encoder.Encoder,
	bcr []byte,
) error {
	e := util.StringErrorFunc("failed to unmarshal CurrencyRegisterFact")

	if hinter, err := enc.Decode(bcr); err != nil {
		return e(err, "")
	} else if cr, ok := hinter.(CurrencyDesign); !ok {
		return util.ErrWrongType.Errorf("expected CurrencyDesign not %T,", hinter)
	} else {
		fact.currency = cr
	}

	return nil
}
