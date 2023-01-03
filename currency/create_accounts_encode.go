package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *CreateAccountsFact) unpack(enc encoder.Encoder, sd string, bit []byte) error {
	e := util.StringErrorFunc("failed to unmarshal CreateAccountsFact")

	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = ad
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e(err, "")
	}

	items := make([]CreateAccountsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateAccountsItem)
		if !ok {
			return util.ErrWrongType.Errorf("expected CreateAccountsItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
