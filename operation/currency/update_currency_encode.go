package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *UpdateCurrencyFact) unpack(enc encoder.Encoder, cid string, bpo []byte) error {
	e := util.StringError("unmarshal UpdateCurrencyFact")

	if hinter, err := enc.Decode(bpo); err != nil {
		return e.Wrap(err)
	} else if po, ok := hinter.(types.CurrencyPolicy); !ok {
		return errors.Errorf("expected CurrencyPolicy, not %T", hinter)
	} else {
		fact.policy = po
	}

	fact.currency = types.CurrencyID(cid)

	return nil
}
