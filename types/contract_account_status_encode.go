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
	oprs []string,
) error {
	cs.BaseHinter = hint.NewBaseHinter(ht)

	switch a, err := base.DecodeAddress(ow, enc); {
	case err != nil:
		return errors.Errorf("Decode address, %v", err)
	default:
		cs.owner = a
	}

	cs.isActive = ia
	operators := make([]base.Address, len(oprs))
	for i, opr := range oprs {
		switch operator, err := base.DecodeAddress(opr, enc); {
		case err != nil:
			return err
		default:
			operators[i] = operator
		}
	}
	cs.operators = operators

	return nil
}
