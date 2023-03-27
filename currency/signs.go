package currency

import (
	"github.com/ProtoconNet/mitum2/base"
)

func checkFactSignsByPubs(pubs []base.Publickey, threshold base.Threshold, signs []base.Sign) error {
	var signed uint
	for i := range signs {
		for j := range pubs {
			if signs[i].Signer().Equal(pubs[j]) {
				signed++

				break
			}
		}
	}

	if float64(signed) < threshold.Float64() {
		return base.NewBaseOperationProcessReasonError("not enough suffrage signs")
	}

	return nil
}

func checkFactSignsByState(
	address base.Address,
	fs []base.Sign,
	getState base.GetStateFunc,
) error {
	st, err := existsState(StateKeyAccount(address), "keys of account", getState)
	if err != nil {
		return err
	}
	keys, err := StateKeysValue(st)
	switch {
	case err != nil:
		return base.NewBaseOperationProcessReasonError("failed to get Keys %w", err)
	case keys == nil:
		return base.NewBaseOperationProcessReasonError("empty keys found")
	}

	if err := checkThreshold(fs, keys); err != nil {
		return base.NewBaseOperationProcessReasonError("failed to check threshold %w", err)
	}

	return nil
}
