package state

import (
	"github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/ProtoconNet/mitum-currency/v2/state/currency"
	mitumbase "github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type StateValueMerger struct {
	*base.BaseStateValueMerger
}

func NewStateValueMerger(height mitumbase.Height, key string, st mitumbase.State) *StateValueMerger {
	s := &StateValueMerger{
		BaseStateValueMerger: base.NewBaseStateValueMerger(height, key, st),
	}

	return s
}

func NewStateMergeValue(key string, stv mitumbase.StateValue) mitumbase.StateMergeValue {
	StateValueMergerFunc := func(height mitumbase.Height, st mitumbase.State) mitumbase.StateValueMerger {
		return NewStateValueMerger(height, key, st)
	}

	return base.NewBaseStateMergeValue(
		key,
		stv,
		StateValueMergerFunc,
	)
}

func CheckNotExistsState(
	key string,
	getState mitumbase.GetStateFunc,
) error {
	switch _, found, err := getState(key); {
	case err != nil:
		return err
	case found:
		return mitumbase.NewBaseOperationProcessReasonError("state, %q already exists", key)
	default:
		return nil
	}
}

func CheckExistsState(
	key string,
	getState mitumbase.GetStateFunc,
) error {
	switch _, found, err := getState(key); {
	case err != nil:
		return err
	case !found:
		return mitumbase.NewBaseOperationProcessReasonError("state, %q does not exist", key)
	default:
		return nil
	}
}

func ExistsState(
	k,
	name string,
	getState mitumbase.GetStateFunc,
) (mitumbase.State, error) {
	switch st, found, err := getState(k); {
	case err != nil:
		return nil, err
	case !found:
		return nil, mitumbase.NewBaseOperationProcessReasonError("%s does not exist", name)
	default:
		return st, nil
	}
}

func NotExistsState(
	k,
	name string,
	getState mitumbase.GetStateFunc,
) (mitumbase.State, error) {
	var st mitumbase.State
	switch _, found, err := getState(k); {
	case err != nil:
		return nil, err
	case found:
		return nil, mitumbase.NewBaseOperationProcessReasonError("%s already exists", name)
	case !found:
		st = base.NewBaseState(mitumbase.NilHeight, k, nil, nil, nil)
	}
	return st, nil
}

func ExistsCurrencyPolicy(cid base.CurrencyID, getStateFunc mitumbase.GetStateFunc) (base.CurrencyPolicy, error) {
	var policy base.CurrencyPolicy
	switch i, found, err := getStateFunc(currency.StateKeyCurrencyDesign(cid)); {
	case err != nil:
		return base.CurrencyPolicy{}, err
	case !found:
		return base.CurrencyPolicy{}, mitumbase.NewBaseOperationProcessReasonError("currency not found, %v", cid)
	default:
		currencydesign, ok := i.Value().(currency.CurrencyDesignStateValue) //nolint:forcetypeassert //...
		if !ok {
			return base.CurrencyPolicy{}, errors.Errorf("expected CurrencyDesignStateValue, not %T", i.Value())
		}
		policy = currencydesign.CurrencyDesign.Policy()
	}
	return policy, nil
}

func CheckFactSignsByState(
	address mitumbase.Address,
	fs []mitumbase.Sign,
	getState mitumbase.GetStateFunc,
) error {
	st, err := ExistsState(currency.StateKeyAccount(address), "keys of account", getState)
	if err != nil {
		return err
	}
	keys, err := currency.StateKeysValue(st)
	switch {
	case err != nil:
		return mitumbase.NewBaseOperationProcessReasonError("failed to get Keys %w", err)
	case keys == nil:
		return mitumbase.NewBaseOperationProcessReasonError("empty keys found")
	}

	if err := base.CheckThreshold(fs, keys); err != nil {
		return mitumbase.NewBaseOperationProcessReasonError("failed to check threshold %w", err)
	}

	return nil
}
