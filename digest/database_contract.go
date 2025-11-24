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
	DefaultColNameContract     = "digest_sc"
	DefaultColNameContractData = "digest_sc_data"
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
