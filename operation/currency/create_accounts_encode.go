package currency

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *CreateAccountsFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringError("failed to unmarshal CreateAccountsFact")

	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.sender = ad
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]CreateAccountsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateAccountsItem)
		if !ok {
			return errors.Errorf("expected CreateAccountsItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
