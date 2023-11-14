package state

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type StateValueMerger struct {
	*common.BaseStateValueMerger
}

func NewStateValueMerger(height base.Height, key string, st base.State) *StateValueMerger {
	s := &StateValueMerger{
		BaseStateValueMerger: common.NewBaseStateValueMerger(height, key, st),
	}

	return s
}

func NewStateMergeValue(key string, stv base.StateValue) base.StateMergeValue {
	StateValueMergerFunc := func(height base.Height, st base.State) base.StateValueMerger {
		nst := st
		if st == nil {
			nst = common.NewBaseState(base.NilHeight, key, nil, nil, nil)
		}
		return NewStateValueMerger(height, nst.Key(), nst)
	}

	return common.NewBaseStateMergeValue(
		key,
		stv,
		StateValueMergerFunc,
	)
}

func CheckNotExistsState(
	key string,
	getState base.GetStateFunc,
) error {
	switch _, found, err := getState(key); {
	case err != nil:
		return err
	case found:
		return base.NewBaseOperationProcessReasonError("state, %v already exists", key)
	default:
		return nil
	}
}

func CheckExistsState(
	key string,
	getState base.GetStateFunc,
) error {
	switch _, found, err := getState(key); {
	case err != nil:
		return err
	case !found:
		return base.NewBaseOperationProcessReasonError("state, %v does not exist", key)
	default:
		return nil
	}
}

func ExistsState(
	k,
	name string,
	getState base.GetStateFunc,
) (base.State, error) {
	switch st, found, err := getState(k); {
	case err != nil:
		return nil, err
	case !found:
		return nil, base.NewBaseOperationProcessReasonError("%v does not exist", name)
	default:
		return st, nil
	}
}

func NotExistsState(
	k,
	name string,
	getState base.GetStateFunc,
) (base.State, error) {
	var st base.State
	switch _, found, err := getState(k); {
	case err != nil:
		return nil, err
	case found:
		return nil, base.NewBaseOperationProcessReasonError("%v already exists", name)
	case !found:
		st = common.NewBaseState(base.NilHeight, k, nil, nil, nil)
	}
	return st, nil
}

func ExistsCurrencyPolicy(cid types.CurrencyID, getStateFunc base.GetStateFunc) (types.CurrencyPolicy, error) {
	var policy types.CurrencyPolicy
	switch i, found, err := getStateFunc(currency.StateKeyCurrencyDesign(cid)); {
	case err != nil:
		return types.CurrencyPolicy{}, err
	case !found:
		return types.CurrencyPolicy{}, base.NewBaseOperationProcessReasonError("currency not found, %v", cid)
	default:
		currencydesign, ok := i.Value().(currency.CurrencyDesignStateValue) //nolint:forcetypeassert //...
		if !ok {
			return types.CurrencyPolicy{}, errors.Errorf("expected CurrencyDesignStateValue, not %T", i.Value())
		}
		policy = currencydesign.CurrencyDesign.Policy()
	}
	return policy, nil
}

func CheckFactSignsByState(
	address base.Address,
	fs []base.Sign,
	getState base.GetStateFunc,
) error {
	st, err := ExistsState(currency.StateKeyAccount(address), "keys of account", getState)
	if err != nil {
		return err
	}
	keys, err := currency.StateKeysValue(st)
	switch {
	case err != nil:
		return base.NewBaseOperationProcessReasonError("get Keys; %w", err)
	case keys == nil:
		return base.NewBaseOperationProcessReasonError("empty keys found")
	}

	if err := types.CheckThreshold(fs, keys); err != nil {
		return base.NewBaseOperationProcessReasonError("check threshold; %w", err)
	}

	return nil
}
