package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *UpdateKeyFact) unpack(enc encoder.Encoder, sd string, bks []byte, cid string) error {
	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return err
	default:
		fact.sender = ad
	}

	if hinter, err := enc.Decode(bks); err != nil {
		return err
	} else if k, ok := hinter.(types.AccountKeys); !ok {
		return common.ErrTypeMismatch.Wrap(errors.Errorf("expected AccountKeys, not %T", hinter))
	} else {
		fact.keys = k
	}

	fact.currency = types.CurrencyID(cid)

	return nil
}
