package extension

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *CreateContractAccountFact) unpack(enc encoder.Encoder, ow string, bit []byte) error {
	e := util.StringError("unmarshal CreateContractAccountFact")

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

	items := make([]CreateContractAccountItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateContractAccountItem)
		if !ok {
			return e.Wrap(errors.Errorf("expected CreateContractAccountItem, not %T", hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
