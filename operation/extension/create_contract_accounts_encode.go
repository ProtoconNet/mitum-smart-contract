package extension

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *CreateContractAccountsFact) unpack(enc encoder.Encoder, ow string, bit []byte) error {
	e := util.StringError("failed to unmarshal CreateContractAccountsFact")

	switch a, err := base.DecodeAddress(ow, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e.Wrap(err)
	}

	items := make([]CreateContractAccountsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateContractAccountsItem)
		if !ok {
			return e.Wrap(errors.Errorf("expected CreateContractAccountsItem, not %T", hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
