package digest

import (
	utilc "github.com/ProtoconNet/mitum-currency/v3/digest/util"
	state "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	types "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	utilm "github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultColNameContract         = "digest_sc"
	DefaultColNameContractData     = "digest_sc_data"
	DefaultColNameContractRuntime  = "digest_sc_runtime"
	DefaultColNameContractSnapshot = "digest_sc_snapshot"
)

func ContractDesign(st *Database, contract string) (types.Design, base.State, error) {
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
			i, err := LoadState(res.Decode, st.Encoders())
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

func ContractData(db *Database, contract, key string) (map[string]interface{}, base.State, error) {
	filter := utilc.NewBSONFilter("contract", contract)
	filter = filter.Add("data_key", key)
	q := filter.D()

	opt := options.FindOne().SetSort(
		utilc.NewBSONFilter("height", -1).D(),
	)
	var data map[string]interface{}
	var sta base.State
	var err error
	if err := db.MongoClient().GetByFilter(
		DefaultColNameContractData,
		q,
		func(res *mongo.SingleResult) error {
			sta, err = LoadState(res.Decode, db.Encoders())
			if err != nil {
				return err
			}
			d, err := state.GetDataFromState(sta)
			if err != nil {
				return err
			}
			data = d
			return nil
		},
		opt,
	); err != nil {
		return nil, nil, utilm.ErrNotFound.WithMessage(
			err, "Contract data for data key %s in contract account %s", key, contract)
	}

	if data != nil {
		return data, sta, nil
	} else {
		return nil, nil, errors.Errorf("data is nil")
	}
}

func ContractDesignFromChainState(db *Database, contract string) (base.Address, types.Design, base.State, error) {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return nil, types.Design{}, nil, errors.Wrap(err, "invalid contract address")
	}

	st, found, err := db.State(state.DesignStateKey(address))
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
	db *Database,
	contract string,
) (base.Address, state.RuntimeStateValue, base.State, bool, error) {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return nil, state.RuntimeStateValue{}, nil, false, errors.Wrap(err, "invalid contract address")
	}

	st, found, err := db.State(state.RuntimeStateKey(address))
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
	db *Database,
	contract string,
) (base.Address, state.SnapshotStateValue, base.State, bool, error) {
	address, err := base.DecodeAddress(contract, db.Encoders().JSON())
	if err != nil {
		return nil, state.SnapshotStateValue{}, nil, false, errors.Wrap(err, "invalid contract address")
	}

	st, found, err := db.State(state.SnapshotStateKey(address))
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
