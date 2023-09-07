package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *UpdateKeyFact) unpack(enc encoder.Encoder, tg string, bks []byte, cid string) error {
	e := util.StringError("unmarshal UpdateKeyFact")

	switch ad, err := base.DecodeAddress(tg, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.target = ad
	}

	if hinter, err := enc.Decode(bks); err != nil {
		return err
	} else if k, ok := hinter.(types.AccountKeys); !ok {
		return errors.Errorf("expected AccountKeys, not %T", hinter)
	} else {
		fact.keys = k
	}

	fact.currency = types.CurrencyID(cid)

	return nil
}
