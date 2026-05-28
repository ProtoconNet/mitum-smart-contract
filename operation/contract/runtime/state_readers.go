package runtime

import (
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
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

type StateBalanceReader struct {
	encs         encoder.Encoders
	getStateFunc base.GetStateFunc
}

func NewStateBalanceReader(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
) StateBalanceReader {
	return StateBalanceReader{
		encs:         encs,
		getStateFunc: getStateFunc,
	}
}

func (r StateBalanceReader) BalanceOf(addr string, currencyID string) (string, bool, error) {
	address, err := base.DecodeAddress(addr, r.encs.JSON())
	if err != nil {
		return "", false, nil
	}

	cid := ctypes.CurrencyID(currencyID)
	if err := cid.IsValid(nil); err != nil {
		return "", false, nil
	}

	accountState, found, err := r.getStateFunc(currency.AccountStateKey(address))
	switch {
	case err != nil:
		return "", false, errors.Errorf("account balance lookup failed for %v: %v", addr, err)
	case !found:
		return "", false, nil
	default:
		if _, err := currency.LoadAccountStateValue(accountState); err != nil {
			return "", false, errors.Wrap(err, "account state decode failed for balance lookup")
		}
	}

	currencyState, found, err := r.getStateFunc(currency.DesignStateKey(cid))
	switch {
	case err != nil:
		return "", false, errors.Errorf("currency balance lookup failed for %v: %v", currencyID, err)
	case !found:
		return "", false, nil
	default:
		if _, err := currency.GetDesignFromState(currencyState); err != nil {
			return "", false, errors.Wrap(err, "currency state decode failed for balance lookup")
		}
	}

	balanceState, found, err := r.getStateFunc(currency.BalanceStateKey(address, cid))
	switch {
	case err != nil:
		return "", false, errors.Errorf("balance lookup failed for %v/%v: %v", addr, currencyID, err)
	case !found:
		return "", false, nil
	}

	amount, err := currency.StateBalanceValue(balanceState)
	if err != nil {
		return "", false, errors.Wrap(err, "balance state decode failed")
	}
	if amount.Currency() != cid {
		return "", false, errors.Errorf("balance state currency mismatch")
	}

	return amount.Big().String(), true, nil
}
