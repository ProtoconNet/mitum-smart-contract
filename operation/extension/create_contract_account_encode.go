package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *CreateContractAccountFact) unpack(enc encoder.Encoder, ow string, bit []byte) error {
	switch a, err := base.DecodeAddress(ow, enc); {
	case err != nil:
		return err
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return err
	}

	items := make([]CreateContractAccountItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateContractAccountItem)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected CreateContractAccountItem, not %T", hit[i]))
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
