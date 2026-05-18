package runtime

import (
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func GetAccountStateFunc(addr string, encs encoder.Encoders, getStateFunc base.GetStateFunc) (bool, error) {
	address, err := base.DecodeAddress(addr, encs.JSON())
	if err != nil {
		return false, errors.Errorf("failed to decode address, %v", addr)
	}

	var st base.State
	var found bool
	k := currency.AccountStateKey(address)
	switch st, found, err = getStateFunc(k); {
	case err != nil:
		return false, errors.Errorf("account, %v: %v", addr, err)
	case !found:
		return false, errors.Errorf("account, %v", addr)
	default:
		_, err = currency.LoadAccountStateValue(st)
		if err != nil {
			return false, errors.Errorf("account, %v: %v", addr, err)
		}
	}

	return true, nil
}
