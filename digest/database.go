package digest

import (
	"context"
	"fmt"
	mongodbstorage "github.com/ProtoconNet/mitum-currency/v3/digest/mongodb"
	"github.com/ProtoconNet/mitum-currency/v3/digest/util"
	"github.com/ProtoconNet/mitum-currency/v3/state/currency"
	"github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	isaacdatabase "github.com/ProtoconNet/mitum2/isaac/database"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var maxLimit int64 = 50

var (
	defaultColNameAccount         = "digest_ac"
	defaultColNameContractAccount = "digest_ca"
	defaultColNameBalance         = "digest_bl"
	defaultColNameCurrency        = "digest_cr"
	defaultColNameOperation       = "digest_op"
	defaultColNameBlock           = "digest_bm"
)

var AllCollections = []string{
	defaultColNameAccount,
	defaultColNameBalance,
	defaultColNameCurrency,
	defaultColNameOperation,
	defaultColNameBlock,
}

var DigestStorageLastBlockKey = "digest_last_block"

type Database struct {
	sync.RWMutex
	*logging.Logging
	mitum     *isaacdatabase.Center
	database  *mongodbstorage.Database
	readonly  bool
	lastBlock base.Height
}

func NewDatabase(mitum *isaacdatabase.Center, st *mongodbstorage.Database) (*Database, error) {
	nst := &Database{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "digest-mongodb-database")
		}),
		mitum:     mitum,
		database:  st,
		lastBlock: base.NilHeight,
	}
	_ = nst.SetLogging(mitum.Logging)

	return nst, nil
}

func NewReadonlyDatabase(mitum *isaacdatabase.Center, st *mongodbstorage.Database) (*Database, error) {
	nst, err := NewDatabase(mitum, st)
	if err != nil {
		return nil, err
	}
	nst.readonly = true

	return nst, nil
}

func (st *Database) New() (*Database, error) {
	if st.readonly {
		return nil, errors.Errorf("readonly mode")
	}

	nst, err := st.database.New()
	if err != nil {
		return nil, err
	}
	return NewDatabase(st.mitum, nst)
}

func (st *Database) Readonly() bool {
	return st.readonly
}

func (st *Database) Close() error {
	return st.database.Close()
}

func (st *Database) DatabaseClient() *mongodbstorage.Client {
	return st.database.Client()
}

func (st *Database) DatabaseEncoder() encoder.Encoder {
	return st.database.Encoder()
}

func (st *Database) DatabaseEncoders() *encoder.Encoders {
	return st.database.Encoders()
}

func (st *Database) Initialize() error {
	st.Lock()
	defer st.Unlock()

	switch h, found, err := loadLastBlock(st); {
	case err != nil:
		return errors.Wrap(err, "initialize digest database")
	case !found:
		st.lastBlock = base.NilHeight
		st.Log().Debug().Msg("last block for digest not found")
	default:
		st.lastBlock = h
	}
	// 	if !st.readonly {
	// 		if err := st.createIndex(); err != nil {
	// 			return err
	// 		}

	// 		if err := st.cleanByHeight(context.Background(), h+1); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	if !st.readonly {
		if err := st.createIndex(); err != nil {
			return err
		}

		// if err := st.cleanByHeight(context.Background(), h+1); err != nil {
		// 	return err
		// }
	}

	return nil
}

func (st *Database) createIndex() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	for col, models := range defaultIndexes {
		if err := st.database.CreateIndex(col, models, indexPrefix); err != nil {
			return err
		}
	}

	return nil
}

func (st *Database) LastBlock() base.Height {
	st.RLock()
	defer st.RUnlock()

	return st.lastBlock
}

func (st *Database) SetLastBlock(height base.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	if height <= st.lastBlock {
		return nil
	}

	return st.setLastBlock(height)
}

func (st *Database) setLastBlock(height base.Height) error {
	if err := st.database.SetInfo(DigestStorageLastBlockKey, height.Bytes()); err != nil {
		st.Log().Debug().Int64("height", height.Int64()).Msg("set last block")

		return err
	}
	st.lastBlock = height
	st.Log().Debug().Int64("height", height.Int64()).Msg("set last block")

	return nil
}

func (st *Database) Clean() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	return st.clean(context.Background())
}

func (st *Database) clean(ctx context.Context) error {
	for _, col := range []string{
		defaultColNameAccount,
		defaultColNameBalance,
		defaultColNameCurrency,
		defaultColNameOperation,
		defaultColNameBlock,
	} {
		if err := st.database.Client().Collection(col).Drop(ctx); err != nil {
			return err
		}

		st.Log().Debug().Str("collection", col).Msg("drop collection by height")
	}

	if err := st.setLastBlock(base.NilHeight); err != nil {
		return err
	}

	st.Log().Debug().Msg("clean digest")

	return nil
}

func (st *Database) CleanByHeight(ctx context.Context, height base.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	return st.cleanByHeight(ctx, height)
}

func (st *Database) cleanByHeight(ctx context.Context, height base.Height) error {
	if height <= base.GenesisHeight {
		return st.clean(ctx)
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByHeight := mongo.NewDeleteManyModel().SetFilter(bson.M{"height": bson.M{"$gte": height}})

	for _, col := range []string{
		defaultColNameAccount,
		defaultColNameBalance,
		defaultColNameCurrency,
		defaultColNameOperation,
		defaultColNameBlock,
	} {
		res, err := st.database.Client().Collection(col).BulkWrite(
			ctx,
			[]mongo.WriteModel{removeByHeight},
			opts,
		)
		if err != nil {
			return err
		}

		st.Log().Debug().Str("collection", col).Interface("result", res).Msg("clean collection by height")
	}

	return st.setLastBlock(height - 1)
}

/*
func (st *Database) Manifest(h mitumutil.Hash) (base.Manifest, bool, error) {
	return st.mitum.Manifest(h)
}
*/

// Manifests returns block.Manifests by order and height.
func (st *Database) Manifests(
	load bool,
	reverse bool,
	offset base.Height,
	limit int64,
	callback func(base.Height, base.Manifest, uint64, string, string, uint64) (bool, error),
) error {
	var filter bson.M
	if offset > base.NilHeight {
		if reverse {
			filter = bson.M{"height": bson.M{"$lt": offset}}
		} else {
			filter = bson.M{"height": bson.M{"$gt": offset}}
		}
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).Add("index", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameBlock,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			va, ops, confirmed, proposer, round, err := LoadManifest(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			return callback(va.Height(), va, ops, confirmed, proposer, round)
		},
		opt,
	)
}

// OperationsByAddress finds the operation.Operations, which are related with
// the given Address. The returned valuehash.Hash is the
// operation.Operation.Fact().Hash().
// *    load:if true, load operation.Operation and returns it. If not, just hash will be returned
// * reverse: order by height; if true, higher height will be returned first.
// *  offset: returns from next of offset, usually it is combination of
// "<height>,<fact>".
func (st *Database) OperationsByAddress(
	address base.Address,
	load,
	reverse bool,
	offset string,
	limit int64,
	callback func(mitumutil.Hash /* fact hash */, OperationValue) (bool, error),
) error {
	filter, err := buildOperationsFilterByAddress(address, offset, reverse)
	if err != nil {
		return err
	}

	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).Add("index", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	if !load {
		opt = opt.SetProjection(bson.M{"fact": 1})
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameOperation,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			if !load {
				h, err := LoadOperationHash(cursor.Decode)
				if err != nil {
					return false, err
				}
				return callback(h, OperationValue{})
			}

			va, err := LoadOperation(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			return callback(va.Operation().Fact().Hash(), va)
		},
		opt,
	)
}

// Operation returns operation.Operation. If load is false, just returns nil
// Operation.
func (st *Database) Operation(
	h mitumutil.Hash, /* fact hash */
	load bool,
) (OperationValue, bool /* exists */, error) {
	if !load {
		exists, err := st.database.Client().Exists(defaultColNameOperation, util.NewBSONFilter("fact", h).D())
		return OperationValue{}, exists, err
	}

	var va OperationValue
	if err := st.database.Client().GetByFilter(
		defaultColNameOperation,
		util.NewBSONFilter("fact", h).D(),
		func(res *mongo.SingleResult) error {
			if !load {
				return nil
			}

			i, err := LoadOperation(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			va = i

			return nil
		},
	); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return OperationValue{}, false, nil
		}

		return OperationValue{}, false, err
	}
	return va, true, nil
}

// Operations returns operation.Operations by order, height and index.
func (st *Database) Operations(
	filter bson.M,
	load bool,
	reverse bool,
	limit int64,
	callback func(mitumutil.Hash /* fact hash */, OperationValue, int64) (bool, error),
) error {
	sr := 1
	if reverse {
		sr = -1
	}

	opt := options.Find().SetSort(
		util.NewBSONFilter("height", sr).Add("index", sr).D(),
	)

	switch {
	case limit <= 0: // no limit
	case limit > maxLimit:
		opt = opt.SetLimit(maxLimit)
	default:
		opt = opt.SetLimit(limit)
	}

	if !load {
		opt = opt.SetProjection(bson.M{"fact": 1})
	}

	count, err := st.database.Client().Count(context.Background(), defaultColNameOperation, bson.D{})
	if err != nil {
		return err
	}

	return st.database.Client().Find(
		context.Background(),
		defaultColNameOperation,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			if !load {
				h, err := LoadOperationHash(cursor.Decode)
				if err != nil {
					return false, err
				}
				return callback(h, OperationValue{}, count)
			}

			va, err := LoadOperation(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}
			return callback(va.Operation().Fact().Hash(), va, count)
		},
		opt,
	)
}

// Account returns AccountValue.
func (st *Database) Account(a base.Address) (AccountValue, bool /* exists */, error) {
	var rs AccountValue
	if err := st.database.Client().GetByFilter(
		defaultColNameAccount,
		util.NewBSONFilter("address", a.String()).D(),
		func(res *mongo.SingleResult) error {
			i, err := LoadAccountValue(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			rs = i

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		//); err != nil {
		//	if errors.Is(err, mitumutil.NewIDError("not found")) {
		//		return rs, false, nil
		//	}

		return rs, false, err
	}

	// NOTE load balance
	switch am, lastHeight, err := st.balance(a); {
	case err != nil:
		return rs, false, err
	default:
		rs = rs.SetBalance(am).
			SetHeight(lastHeight)
	}
	// NOTE load contract account status
	switch status, lastHeight, err := st.contractAccountStatus(a); {
	case err != nil:
		return rs, true, nil
	default:
		rs = rs.SetContractAccountStatus(status).
			SetHeight(lastHeight)
	}

	return rs, true, nil
}

// AccountsByPublickey finds Accounts, which are related with the given
// Publickey.
// *  offset: returns from next of offset, usually it is "<height>,<address>".
func (st *Database) AccountsByPublickey(
	pub base.Publickey,
	loadBalance bool,
	offsetHeight base.Height,
	offsetAddress string,
	limit int64,
	callback func(AccountValue) (bool, error),
) error {
	if offsetHeight <= base.NilHeight {
		return errors.Errorf("offset height should be over nil height")
	}

	filter := buildAccountsFilterByPublickey(pub)
	filter["height"] = bson.M{"$lte": offsetHeight}

	var sas []string
	switch i, err := st.addressesByPublickey(filter); {
	case err != nil:
		return err
	default:
		sas = i
	}

	if len(sas) < 1 {
		return nil
	}

	var filteredAddress []string
	if len(offsetAddress) < 1 {
		filteredAddress = sas
	} else {
		var found bool
		for i := range sas {
			a := sas[i]
			if !found {
				if offsetAddress == a {
					found = true
				}

				continue
			}

			filteredAddress = append(filteredAddress, a)
		}
	}

	if len(filteredAddress) < 1 {
		return nil
	}

end:
	for i := int64(0); i < int64(math.Ceil(float64(len(filteredAddress))/50.0)); i++ {
		l := (i + 1) + 50
		if n := int64(len(filteredAddress)); l > n {
			l = n
		}

		limited := filteredAddress[i*50 : l]
		switch done, err := st.filterAccountByPublickey(
			pub, limited, limit, loadBalance, callback,
		); {
		case err != nil:
			return err
		case done:
			break end
		}
	}

	return nil
}

func (st *Database) balance(a base.Address) ([]types.Amount, base.Height, error) {
	lastHeight := base.NilHeight
	var cids []string

	amm := map[types.CurrencyID]types.Amount{}
	for {
		filter := util.NewBSONFilter("address", a.String())

		var q primitive.D
		if len(cids) < 1 {
			q = filter.D()
		} else {
			q = filter.Add("currency", bson.M{"$nin": cids}).D()
		}

		var sta base.State
		if err := st.database.Client().GetByFilter(
			defaultColNameBalance,
			q,
			func(res *mongo.SingleResult) error {
				i, err := LoadBalance(res.Decode, st.database.Encoders())
				if err != nil {
					return err
				}
				sta = i

				return nil
			},
			options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
		); err != nil {
			if err.Error() == mitumutil.NewIDError("mongo: no documents in result").Error() {
				break
			}

			return nil, lastHeight, err
		}

		i, err := currency.StateBalanceValue(sta)
		if err != nil {
			return nil, lastHeight, err
		}
		amm[i.Currency()] = i

		cids = append(cids, i.Currency().String())

		if h := sta.Height(); h > lastHeight {
			lastHeight = h
		}
	}

	ams := make([]types.Amount, len(amm))
	var i int
	for k := range amm {
		ams[i] = amm[k]
		i++
	}

	return ams, lastHeight, nil
}

func (st *Database) contractAccountStatus(a base.Address) (types.ContractAccountStatus, base.Height, error) {
	lastHeight := base.NilHeight

	filter := util.NewBSONFilter("address", a)
	filter.Add("contract", true)

	opt := options.FindOne().SetSort(
		util.NewBSONFilter("height", -1).D(),
	)
	var sta base.State
	if err := st.database.Client().GetByFilter(
		defaultColNameContractAccount,
		filter.D(),
		func(res *mongo.SingleResult) error {
			i, err := LoadContractAccountStatus(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			sta = i
			return nil
		},
		opt,
	); err != nil {
		return types.ContractAccountStatus{}, lastHeight, err
	}

	if sta != nil {
		cas, err := extension.StateContractAccountValue(sta)
		if err != nil {
			return types.ContractAccountStatus{}, lastHeight, err
		}
		if h := sta.Height(); h > lastHeight {
			lastHeight = h
		}
		return cas, lastHeight, nil
	} else {
		return types.ContractAccountStatus{}, lastHeight, errors.Errorf("state is nil")
	}
}

func (st *Database) currencies() ([]string, error) {
	var cids []string

	for {
		filter := util.EmptyBSONFilter()

		var q primitive.D
		if len(cids) < 1 {
			q = filter.D()
		} else {
			q = filter.Add("currency", bson.M{"$nin": cids}).D()
		}

		opt := options.FindOne().SetSort(
			util.NewBSONFilter("height", -1).D(),
		)
		var sta base.State
		if err := st.database.Client().GetByFilter(
			defaultColNameCurrency,
			q,
			func(res *mongo.SingleResult) error {
				i, err := LoadCurrency(res.Decode, st.database.Encoders())
				if err != nil {
					return err
				}
				sta = i
				return nil
			},
			opt,
		); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				break
			}

			return nil, err
		}

		if sta != nil {
			i, err := currency.StateCurrencyDesignValue(sta)
			if err != nil {
				return nil, err
			}
			cids = append(cids, i.Currency().String())
		} else {
			return nil, errors.Errorf("state is nil")
		}

	}

	return cids, nil
}

func (st *Database) ManifestByHeight(height base.Height) (base.Manifest, uint64, string, string, uint64, error) {
	q := util.NewBSONFilter("height", height).D()

	var m base.Manifest
	var operations, round uint64
	var confirmed, proposer string
	if err := st.database.Client().GetByFilter(
		defaultColNameBlock,
		q,
		func(res *mongo.SingleResult) error {
			v, ops, cfrm, prps, rnd, err := LoadManifest(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			m = v
			operations = ops
			confirmed = cfrm
			proposer = prps
			round = rnd
			return nil
		},
	); err != nil {
		return nil, 0, "", "", 0, mitumutil.ErrNotFound.WithMessage(err, "block manifest")
	}

	if m != nil {
		return m, operations, confirmed, proposer, round, nil
	} else {
		return nil, 0, "", "", 0, mitumutil.ErrNotFound.Wrap(errors.Errorf("block manifest"))
	}
}

func (st *Database) ManifestByHash(hash mitumutil.Hash) (base.Manifest, uint64, string, string, uint64, error) {
	q := util.NewBSONFilter("block", hash).D()

	var m base.Manifest
	var operations, round uint64
	var confirmed, proposer string
	if err := st.database.Client().GetByFilter(
		defaultColNameBlock,
		q,
		func(res *mongo.SingleResult) error {
			v, ops, cfrm, prps, rnd, err := LoadManifest(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			m = v
			operations = ops
			confirmed = cfrm
			proposer = prps
			round = rnd
			return nil
		},
	); err != nil {
		return nil, 0, "", "", 0, mitumutil.ErrNotFound.WithMessage(err, "block manifest")
	}

	if m != nil {
		return m, operations, confirmed, proposer, round, nil
	} else {
		return nil, 0, "", "", 0, mitumutil.ErrNotFound.Errorf("block manifest")
	}
}

func (st *Database) currency(cid string) (types.CurrencyDesign, base.State, error) {
	q := util.NewBSONFilter("currency", cid).D()

	opt := options.FindOne().SetSort(
		util.NewBSONFilter("height", -1).D(),
	)
	var sta base.State
	if err := st.database.Client().GetByFilter(
		defaultColNameCurrency,
		q,
		func(res *mongo.SingleResult) error {
			i, err := LoadCurrency(res.Decode, st.database.Encoders())
			if err != nil {
				return err
			}
			sta = i
			return nil
		},
		opt,
	); err != nil {
		return types.CurrencyDesign{}, nil, mitumutil.ErrNotFound.WithMessage(err, "currency in handleCurrency")
	}

	if sta != nil {
		de, err := currency.StateCurrencyDesignValue(sta)
		if err != nil {
			return types.CurrencyDesign{}, nil, err
		}
		return de, sta, nil
	} else {
		return types.CurrencyDesign{}, nil, errors.Errorf("state is nil")
	}
}

func (st *Database) topHeightByPublickey(pub base.Publickey) (base.Height, error) {
	var sas []string
	switch r, err := st.database.Client().Collection(defaultColNameAccount).Distinct(
		context.Background(),
		"address",
		buildAccountsFilterByPublickey(pub),
	); {
	case err != nil:
		return base.NilHeight, err
	case len(r) < 1:
		return base.NilHeight, err
	default:
		sas = make([]string, len(r))
		for i := range r {
			sas[i] = r[i].(string)
		}
	}

	var top base.Height
	for i := int64(0); i < int64(math.Ceil(float64(len(sas))/50.0)); i++ {
		l := (i + 1) + 50
		if n := int64(len(sas)); l > n {
			l = n
		}

		switch h, err := st.partialTopHeightByPublickey(sas[i*50 : l]); {
		case err != nil:
			return base.NilHeight, err
		case top <= base.NilHeight:
			top = h
		case h > top:
			top = h
		}
	}

	return top, nil
}

func (st *Database) partialTopHeightByPublickey(as []string) (base.Height, error) {
	var top base.Height
	err := st.database.Client().Find(
		context.Background(),
		defaultColNameAccount,
		bson.M{"address": bson.M{"$in": as}},
		func(cursor *mongo.Cursor) (bool, error) {
			h, err := loadHeightDoc(cursor.Decode)
			if err != nil {
				return false, err
			}

			top = h

			return false, nil
		},
		options.Find().
			SetSort(util.NewBSONFilter("height", -1).D()).
			SetLimit(1),
	)

	return top, err
}

func (st *Database) addressesByPublickey(filter bson.M) ([]string, error) {
	r, err := st.database.Client().Collection(defaultColNameAccount).Distinct(context.Background(), "address", filter)
	if err != nil {
		return nil, errors.Wrap(err, "get distinct addresses")
	}

	if len(r) < 1 {
		return nil, nil
	}

	sas := make([]string, len(r))
	for i := range r {
		sas[i] = r[i].(string)
	}

	sort.Strings(sas)

	return sas, nil
}

func (st *Database) filterAccountByPublickey(
	pub base.Publickey,
	addresses []string,
	limit int64,
	loadBalance bool,
	callback func(AccountValue) (bool, error),
) (bool, error) {
	filter := bson.M{"address": bson.M{"$in": addresses}}

	var lastAddress string
	var called int64
	var stopped bool
	if err := st.database.Client().Find(
		context.Background(),
		defaultColNameAccount,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			if called == limit {
				return false, nil
			}

			doc, err := loadBriefAccountDoc(cursor.Decode)
			if err != nil {
				return false, err
			}

			if len(lastAddress) > 0 {
				if lastAddress == doc.Address {
					return true, nil
				}
			}
			lastAddress = doc.Address

			if !doc.pubExists(pub) {
				return true, nil
			}

			va, err := LoadAccountValue(cursor.Decode, st.database.Encoders())
			if err != nil {
				return false, err
			}

			if loadBalance { // NOTE load balance
				switch am, lastHeight, err := st.balance(va.Account().Address()); {
				case err != nil:
					return false, err
				default:
					va = va.SetBalance(am).
						SetHeight(lastHeight)
				}
			}

			called++
			switch keep, err := callback(va); {
			case err != nil:
				return false, err
			case !keep:
				stopped = true

				return false, nil
			default:
				return true, nil
			}
		},
		options.Find().SetSort(util.NewBSONFilter("address", 1).Add("height", -1).D()),
	); err != nil {
		return false, err
	}

	return stopped || called == limit, nil
}

func (st *Database) CleanByHeightColName(
	ctx context.Context,
	height base.Height,
	colName, key, value string,
) error {
	if height <= base.GenesisHeight {
		return st.clean(ctx)
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByHeight := mongo.NewDeleteManyModel().SetFilter(
		bson.M{key: value, "height": bson.M{"$lte": height}},
	)

	res, err := st.database.Client().Collection(colName).BulkWrite(
		ctx,
		[]mongo.WriteModel{removeByHeight},
		opts,
	)
	if err != nil {
		return err
	}

	st.Log().Debug().Str("collection", colName).Interface("result", res).Msg("clean collection by height")

	return st.setLastBlock(height - 1)
}

func (st *Database) cleanBalanceByHeightAndAccount(ctx context.Context, height base.Height, address string) error {
	if height <= base.GenesisHeight+1 {
		return st.clean(ctx)
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByAddress := mongo.NewDeleteManyModel().SetFilter(bson.M{"address": address, "height": bson.M{"$lte": height}})

	res, err := st.database.Client().Collection(defaultColNameBalance).BulkWrite(
		context.Background(),
		[]mongo.WriteModel{removeByAddress},
		opts,
	)
	if err != nil {
		return err
	}

	st.Log().Debug().Str("collection", defaultColNameBalance).Interface("result", res).Msg("clean Balancecollection by address")

	return st.setLastBlock(height - 1)
}

func loadLastBlock(st *Database) (base.Height, bool, error) {
	switch b, found, err := st.database.Info(DigestStorageLastBlockKey); {
	case err != nil:
		return base.NilHeight, false, errors.Wrap(err, "get last block for digest")
	case !found:
		return base.NilHeight, false, nil
	default:
		h, err := base.ParseHeightBytes(b)
		if err != nil {
			return base.NilHeight, false, err
		}
		return h, true, nil
	}
}

func parseOffset(s string) (base.Height, uint64, error) {
	if n := strings.SplitN(s, ",", 2); n == nil {
		return base.NilHeight, 0, errors.Errorf("invalid offset string: %q", s)
	} else if len(n) < 2 {
		return base.NilHeight, 0, errors.Errorf("invalid offset, %q", s)
	} else if h, err := base.ParseHeightString(n[0]); err != nil {
		return base.NilHeight, 0, errors.Wrap(err, "invalid height of offset")
	} else if u, err := strconv.ParseUint(n[1], 10, 64); err != nil {
		return base.NilHeight, 0, errors.Wrap(err, "invalid index of offset")
	} else {
		return h, u, nil
	}
}

func buildOffset(height base.Height, index uint64) string {
	return fmt.Sprintf("%d,%d", height, index)
}

func buildOperationsFilterByAddress(address base.Address, offset string, reverse bool) (bson.M, error) {
	filter := bson.M{"addresses": bson.M{"$in": []string{address.String()}}}
	if len(offset) > 0 {
		height, index, err := parseOffset(offset)
		if err != nil {
			return nil, err
		}

		if reverse {
			filter["$or"] = []bson.M{
				{"height": bson.M{"$lt": height}},
				{"$and": []bson.M{
					{"height": height},
					{"index": bson.M{"$lt": index}},
				}},
			}
		} else {
			filter["$or"] = []bson.M{
				{"height": bson.M{"$gt": height}},
				{"$and": []bson.M{
					{"height": height},
					{"index": bson.M{"$gt": index}},
				}},
			}
		}
	}

	return filter, nil
}

func parseOffsetByString(s string) (base.Height, string, error) {
	var a, b string
	switch n := strings.SplitN(s, ",", 2); {
	case n == nil:
		return base.NilHeight, "", errors.Errorf("invalid offset string: %q", s)
	case len(n) < 2:
		return base.NilHeight, "", errors.Errorf("invalid offset, %q", s)
	default:
		a = n[0]
		b = n[1]
	}

	h, err := base.ParseHeightString(a)
	if err != nil {
		return base.NilHeight, "", errors.Wrap(err, "invalid height of offset")
	}

	return h, b, nil
}

func buildOffsetByString(height base.Height, s string) string {
	return fmt.Sprintf("%d,%s", height, s)
}

func buildAccountsFilterByPublickey(pub base.Publickey) bson.M {
	return bson.M{"pubs": bson.M{"$in": []string{pub.String()}}}
}

type heightDoc struct {
	H base.Height `bson:"height"`
}

func loadHeightDoc(decoder func(interface{}) error) (base.Height, error) {
	var h heightDoc
	if err := decoder(&h); err != nil {
		return base.NilHeight, err
	}

	return h.H, nil
}

type briefAccountDoc struct {
	ID      primitive.ObjectID `bson:"_id"`
	Address string             `bson:"address"`
	Pubs    []string           `bson:"pubs"`
	Height  base.Height        `bson:"height"`
}

func (doc briefAccountDoc) pubExists(k base.PKKey) bool {
	if len(doc.Pubs) < 1 {
		return false
	}

	for i := range doc.Pubs {
		if k.String() == doc.Pubs[i] {
			return true
		}
	}

	return false
}

func loadBriefAccountDoc(decoder func(interface{}) error) (briefAccountDoc, error) {
	var a briefAccountDoc
	if err := decoder(&a); err != nil {
		return a, err
	}

	return a, nil
}
