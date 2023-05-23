package extension

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CreateContractAccountsFact) unpack(enc encoder.Encoder, ow string, bit []byte) error {
	e := util.StringErrorFunc("failed to unmarshal CreateContractAccountsFact")

	switch a, err := base.DecodeAddress(ow, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(bit)
	if err != nil {
		return e(err, "")
	}

	items := make([]CreateContractAccountsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateContractAccountsItem)
		if !ok {
			return e(util.ErrWrongType.Errorf("expected CreateContractAccountsItem, not %T", hit[i]), "")
		}

		items[i] = j
	}
	fact.items = items

	return nil
}
