package digest

import (
	"reflect"
	"strings"
	"unsafe"

	cdigest "github.com/imfact-labs/currency-model/digest"
	utilc "github.com/imfact-labs/currency-model/digest/util"
	"github.com/imfact-labs/mitum2/base"
	utilm "github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/smart-contract-model/state"
	"github.com/imfact-labs/smart-contract-model/types"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	defaultColNameAccount          = "digest_ac"
	defaultColNameContractAccount  = "digest_ca"
	defaultColNameBalance          = "digest_bl"
	defaultColNameCurrency         = "digest_cr"
	defaultColNameOperation        = "digest_op"
	DefaultColNameContract         = "digest_sc"
	DefaultColNameContractRuntime  = "digest_sc_runtime"
	DefaultColNameContractSnapshot = "digest_sc_snapshot"
	defaultColNameBlock            = "digest_bm"
)

func ContractDesign(st *cdigest.Database, contract string) (types.Design, base.State, error) {
	filter := utilc.NewBSONFilter("contract", contract)
	q := filter.D()

	opt := options.FindOne().SetSort(
		utilc.NewBSONFilter("height", -1).D(),
	)
	var sta base.State
	if err := st.MongoClient().GetByFilter(
		DefaultColNameContract,
		q,
		func(res *mongo.SingleResult) error {
			i, err := cdigest.LoadState(res.Decode, st.Encoders())
			if err != nil {
				return err
			}
			sta = i
			return nil
		},
		opt,
	); err != nil {
		return types.Design{}, nil, utilm.ErrNotFound.WithMessage(err, "Contract design by contract account %v", contract)
	}

	if sta != nil {
		de, err := state.GetDesignFromState(sta)
		if err != nil {
			return types.Design{}, nil, err
		}
		return de, sta, nil
	} else {
		return types.Design{}, nil, errors.Errorf("state is nil")
	}
}

func ContractDesignFromChainState(db *cdigest.Database, contract string) (base.Address, types.Design, base.State, error) {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return nil, types.Design{}, nil, errors.Wrap(err, "invalid contract address")
	}

	st, found, err := latestContractState(db, DefaultColNameContract, contract)
	if err != nil {
		return nil, types.Design{}, nil, errors.Wrap(err, "failed to read design state from chain")
	}
	if !found {
		return nil, types.Design{}, nil, utilm.ErrNotFound.Errorf("contract design not found for %s", contract)
	}

	de, err := state.GetDesignFromState(st)
	if err != nil {
		return nil, types.Design{}, nil, err
	}

	return address, de, st, nil
}

func ContractRuntimeFromChainState(
	db *cdigest.Database,
	contract string,
) (base.Address, state.RuntimeStateValue, base.State, bool, error) {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return nil, state.RuntimeStateValue{}, nil, false, errors.Wrap(err, "invalid contract address")
	}

	st, found, err := latestContractState(db, DefaultColNameContractRuntime, contract)
	if err != nil {
		return nil, state.RuntimeStateValue{}, nil, false, errors.Wrap(err, "failed to read runtime state from chain")
	}
	if !found {
		return address, state.RuntimeStateValue{}, nil, false, nil
	}

	rv, err := state.GetRuntimeFromState(st)
	if err != nil {
		return nil, state.RuntimeStateValue{}, nil, false, err
	}

	return address, rv, st, true, nil
}

func ContractSnapshotFromChainState(
	db *cdigest.Database,
	contract string,
) (base.Address, state.SnapshotStateValue, base.State, bool, error) {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return nil, state.SnapshotStateValue{}, nil, false, errors.Wrap(err, "invalid contract address")
	}

	st, found, err := latestContractState(db, DefaultColNameContractSnapshot, contract)
	if err != nil {
		return nil, state.SnapshotStateValue{}, nil, false, errors.Wrap(err, "failed to read snapshot state from chain")
	}
	if !found {
		return address, state.SnapshotStateValue{}, nil, false, nil
	}

	sv, err := state.GetSnapshotFromState(st)
	if err != nil {
		return nil, state.SnapshotStateValue{}, nil, false, err
	}

	return address, sv, st, true, nil
}

func latestContractState(db *cdigest.Database, collection, contract string) (base.State, bool, error) {
	if db.MongoClient() == nil {
		return chainState(db, contractStateKey(db, collection, contract))
	}

	filter := utilc.NewBSONFilter("contract", contract)
	opt := options.FindOne().SetSort(utilc.NewBSONFilter("height", -1).D())

	var st base.State
	if err := db.MongoClient().GetByFilter(
		collection,
		filter.D(),
		func(res *mongo.SingleResult) error {
			i, err := cdigest.LoadState(res.Decode, db.Encoders())
			if err != nil {
				return err
			}
			st = i
			return nil
		},
		opt,
	); err != nil {
		return nil, false, err
	}

	return st, st != nil, nil
}

func chainState(db *cdigest.Database, key string) (base.State, bool, error) {
	v := reflect.ValueOf(db)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, false, utilm.ErrNotFound.Errorf("state not found for %s", key)
	}

	field := v.Elem().FieldByName("mitumDB")
	if !field.IsValid() || field.IsNil() {
		return nil, false, utilm.ErrNotFound.Errorf("state not found for %s", key)
	}

	i := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
	stater, ok := i.(interface {
		State(string) (base.State, bool, error)
	})
	if !ok {
		return nil, false, utilm.ErrNotFound.Errorf("state not found for %s", key)
	}

	return stater.State(key)
}

func contractStateKey(db *cdigest.Database, collection, contract string) string {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return ""
	}

	switch collection {
	case DefaultColNameContract:
		return state.DesignStateKey(address)
	case DefaultColNameContractRuntime:
		return state.RuntimeStateKey(address)
	case DefaultColNameContractSnapshot:
		return state.SnapshotStateKey(address)
	default:
		return ""
	}
}

func contractDigestGetStateFunc(db *cdigest.Database) base.GetStateFunc {
	return func(key string) (base.State, bool, error) {
		contract, err := contractFromStateKey(key)
		if err != nil {
			return nil, false, err
		}

		switch {
		case state.IsDesignStateKey(key):
			return latestContractState(db, DefaultColNameContract, contract)
		case state.IsRuntimeStateKey(key):
			return latestContractState(db, DefaultColNameContractRuntime, contract)
		case state.IsSnapshotStateKey(key):
			return latestContractState(db, DefaultColNameContractSnapshot, contract)
		default:
			return nil, false, utilm.ErrNotFound.Errorf("state not found for %s", key)
		}
	}
}

func contractFromStateKey(key string) (string, error) {
	parts := strings.Split(key, ":")
	if len(parts) < 3 || parts[0] != state.ContractStateKeyPrefix {
		return "", errors.Errorf("invalid contract state key %q", key)
	}

	return parts[1], nil
}
