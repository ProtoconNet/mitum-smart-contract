package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *CreateAccountFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return err
	default:
		fact.sender = ad
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return err
	}

	items := make([]CreateAccountItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateAccountItem)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected CreateAccountItem, not %T", hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
