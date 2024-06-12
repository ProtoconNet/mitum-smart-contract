package types // nolint: dupl, revive

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (cs *ContractAccountStatus) unpack(
	enc encoder.Encoder,
	ht hint.Hint,
	ow string,
	ia bool,
	hds []string,
) error {
	cs.BaseHinter = hint.NewBaseHinter(ht)

	switch a, err := base.DecodeAddress(ow, enc); {
	case err != nil:
		return errors.Errorf("Decode address, %v", err)
	default:
		cs.owner = a
	}

	cs.isActive = ia
	handlers := make([]base.Address, len(hds))
	for i, opr := range hds {
		switch handler, err := base.DecodeAddress(opr, enc); {
		case err != nil:
			return err
		default:
			handlers[i] = handler
		}
	}
	cs.handlers = handlers

	return nil
}
