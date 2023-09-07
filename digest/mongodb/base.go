package mongodbstorage

import (
	"context"
	"sync"
	"time"

	"github.com/ProtoconNet/mitum-currency/v3/digest/util"
	bsonenc "github.com/ProtoconNet/mitum-currency/v3/digest/util/bson"
	"github.com/ProtoconNet/mitum2/base"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ColNameInfo     = "info"
	ColNameManifest = "manifest"
	// ColNameOperation = "operation"
	// ColNameStagedOperation = "staged_operation"
	// ColNameProposal     = "proposal"
	// ColNameState        = "state"
	// ColNameVoteproof    = "voteproof"
	ColNameBlockdataMap = "blockdata_map"
)

var allCollections = []string{
	ColNameInfo,
	ColNameManifest,
	// ColNameOperation,
	// ColNameStagedOperation,
	// ColNameProposal,
	// ColNameState,
	// ColNameVoteproof,
	ColNameBlockdataMap,
}

type Database struct {
	sync.RWMutex
	*logging.Logging
	client              *Client
	encs                *encoder.Encoders
	enc                 encoder.Encoder
	lastManifest        base.Manifest
	lastManifestHeight  base.Height
	readonly            bool
	cache               gcache.Cache
	lastINITVoteproof   base.Voteproof
	lastACCEPTVoteproof base.Voteproof
}

func NewDatabase(client *Client, encs *encoder.Encoders, enc encoder.Encoder) (*Database, error) {
	// NOTE call Initialize() later.
	if enc == nil {
		e, found := encs.Find(bsonenc.BSONEncoderHint)
		if !found {
			return nil, mitumutil.ErrNotFound.Errorf("unknown encoder hint, %q", bsonenc.BSONEncoderHint)
		} else {
			enc = e
		}
	}

	return &Database{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "mongodb-database")
		}),
		client:             client,
		encs:               encs,
		enc:                enc,
		lastManifestHeight: base.NilHeight,
	}, nil
}

func NewDatabaseFromURI(uri string, encs *encoder.Encoders) (*Database, error) {
	parsed, err := util.ParseURL(uri, false)
	if err != nil {
		return nil, errors.Wrap(err, "invalid storge uri")
	}

	connectTimeout := time.Second * 7
	execTimeout := time.Second * 7
	{
		query := parsed.Query()
		if d, err := parseDurationFromQuery(query, "connectTimeout", connectTimeout); err != nil {
			return nil, err
		} else {
			connectTimeout = d
		}
		if d, err := parseDurationFromQuery(query, "execTimeout", execTimeout); err != nil {
			return nil, err
		} else {
			execTimeout = d
		}
	}

	var be encoder.Encoder
	if e, found := encs.Find(bsonenc.BSONEncoderHint); !found { // NOTE get latest bson encoder
		return nil, mitumutil.ErrNotFound.Errorf("unknown encoder hint, %q", bsonenc.BSONEncoderHint)
	} else {
		be = e
	}

	if client, err := NewClient(uri, connectTimeout, execTimeout); err != nil {
		return nil, err
	} else if st, err := NewDatabase(client, encs, be); err != nil {
		return nil, err
	} else {
		return st, nil
	}
}

func (st *Database) Initialize() error {
	if st.readonly {
		st.lastManifestHeight = base.Height(int(^uint(0) >> 1))

		return nil
	}

	// if err := st.loadLastBlock(); err != nil && !errors.Is(err, mitumutil.ErrNotFound) {
	// 	return err
	// }
	/*
		if err := st.cleanupIncompleteData(); err != nil {
			return err
		}
	*/

	return st.initialize()
}

/*
func (st *Database) loadLastBlock() error {
	var height base.Height
	if err := st.client.GetByID(ColNameInfo, lastManifestDocID,
		func(res *mongo.SingleResult) error {
			if i, err := loadLastManifest(res.Decode, st.encs); err != nil {
				return err
			} else {
				height = i
			}

			return nil
		},
	); err != nil {
		return err
	}

	switch m, found, err := st.manifestByFilter(util.NewBSONFilter("height", height).D()); {
	case err != nil:
		return errors.Wrapf(err, "find last block of height, %v", height)
	case !found:
		return mitumutil.ErrNotFound.Errorf("find last block of height, %v", height)
	default:
		return st.setLastBlock(m, false, false)
	}
}

func (st *Database) SaveLastBlock(height base.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if cb, err := NewLastManifestDoc(height, st.enc); err != nil {
		return err
	} else if _, err := st.client.Set(ColNameInfo, cb); err != nil {
		return err
	}

	return nil
}

func (st *Database) lastHeight() base.Height {
	st.RLock()
	defer st.RUnlock()

	return st.lastManifestHeight
}


func (st *Database) LastManifest() (base.Manifest, bool, error) {
	if st.readonly {
		return st.manifestByFilter(bson.D{})
	}

	st.RLock()
	defer st.RUnlock()

	if st.lastManifest == nil {
		return nil, false, nil
	}

	return st.lastManifest, true, nil
}


func (st *Database) setLastManifest(manifest base.Manifest, save, force bool) error {
	st.Lock()
	defer st.Unlock()

	return st.setLastManifestInternal(manifest, save, force)
}

func (st *Database) setLastManifestInternal(manifest block.Manifest, save, force bool) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if manifest == nil {
		if save {
			if err := st.SaveLastBlock(base.NilHeight); err != nil {
				return err
			}
		}

		st.lastManifest = nil
		st.lastManifestHeight = base.GenesisHeight

		return nil
	}

	if !force && manifest.Height() <= st.lastManifestHeight {
		return nil
	}

	if save {
		if err := st.SaveLastBlock(manifest.Height()); err != nil {
			return err
		}
	}

	st.Log().Debug().Int64("block_height", manifest.Height().Int64()).Msg("new last block")

	switch t := manifest.(type) {
	case block.Block:
		manifest = t.Manifest()
	}

	st.lastManifest = manifest
	st.lastManifestHeight = manifest.Height()

	return nil
}

func (st *Database) setLastBlock(manifest base.Manifest, save, force bool) error {
	st.Lock()
	defer st.Unlock()

	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	lastManifestHeight := st.lastManifestHeight
	if err := st.setLastManifestInternal(manifest, save, force); err != nil {
		return err
	}

	if manifest == nil {
		st.lastINITVoteproof = nil
		st.lastACCEPTVoteproof = nil

		return nil
	}

	if !force && manifest.Height() <= lastManifestHeight {
		return nil
	}

	var initVoteproof, acceptVoteproof base.Voteproof
	if manifest.Height() > base.GenesisHeight {
		if i, j, err := st.lastVoteproofs(manifest.Height()); err != nil {
			return err
		} else {
			initVoteproof = i
			acceptVoteproof = j
		}
	}

	st.lastINITVoteproof = initVoteproof
	st.lastACCEPTVoteproof = acceptVoteproof

	return nil
}
*/

func (st *Database) Client() *Client {
	return st.client
}

func (st *Database) Close() error {
	// FUTURE return st.client.Close()
	return nil
}

/*
// Clean will drop the existing collections. To keep safe the another
// collections by user, drop collections instead of drop database.
func (st *Database) Clean() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	drop := func(c string) error {
		return st.client.Collection(c).Drop(context.Background())
	}

	for _, c := range allCollections {
		if err := drop(c); err != nil {
			return err
		}
	}

	if err := st.initialize(); err != nil {
		return err
	}

	st.Lock()
	defer st.Unlock()

	st.lastManifest = nil
	st.lastManifestHeight = base.NilHeight
	st.lastINITVoteproof = nil
	st.lastACCEPTVoteproof = nil

	return nil
}

func (st *Database) CleanByHeight(height base.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if err := st.cleanByHeight(height); err != nil {
		return err
	} else if height <= base.GenesisHeight {
		return nil
	}

	switch m, found, err := st.LastManifest(); {
	case err != nil:
		return err
	case !found:
		//
	case m.Height() == height-1:
		return nil
	}

	switch m, found, err := st.ManifestByHeight(height - 1); {
	case err != nil:
		return errors.Wrapf(err, "find block of height, %v", height-1)
	case !found:
		return mitumutil.ErrNotFound.Errorf("find block of height, %v", height-1)
	default:
		_ = st.stateCache.Purge()
		_ = st.operationFactCache.Purge()

		return st.setLastBlock(m, true, true)
	}
}
*/

func (st *Database) Encoder() encoder.Encoder {
	return st.enc
}

func (st *Database) Encoders() *encoder.Encoders {
	return st.encs
}

/*
func (st *Database) Cache() gcache.Cache {
	return st.cache
}
*/

/*
func (st *Database) manifestByFilter(filter bson.D) (block.Manifest, bool, error) {
	var manifest block.Manifest

	if err := st.client.GetByFilter(
		ColNameManifest,
		filter,
		func(res *mongo.SingleResult) error {
			if i, err := loadManifestFromDecoder(res.Decode, st.encs); err != nil {
				return err
			} else {
				manifest = i
			}

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		if errors.Is(err, mitumutil.ErrNotFound) {
			return nil, false, nil
		}

		return nil, false, err
	}

	if manifest == nil {
		return nil, false, nil
	}

	return manifest, true, nil
}

func (st *Database) Manifest(h mitumutil.Hash) (base.Manifest, bool, error) {
	switch m, found, err := st.LastManifest(); {
	case err != nil:
		return nil, false, err
	case found && m.Hash().Equal(h):
		return m, true, nil
	}

	return st.manifestByFilter(util.NewBSONFilter("_id", h.String()).AddOp("height", st.lastHeight(), "$lte").D())
}

func (st *Database) ManifestByHeight(height base.Height) (base.Manifest, bool, error) {
	switch m, found, err := st.LastManifest(); {
	case err != nil:
		return nil, false, err
	case found && m.Height() == height:
		return m, true, nil
	}

	return st.manifestByFilter(util.NewBSONFilter("height", height).AddOp("height", st.lastHeight(), "$lte").D())
}

func (st *Database) Manifests(load, reverse bool, limit int64, callback func(base.Height, valuehash.Hash, base.Manifest) (bool, error)) error {
	return st.ManifestsByFilter(bson.D{}, load, reverse, limit, callback)
}

func (st *Database) ManifestsByFilter(filter interface{}, load, reverse bool, limit int64, callback func(base.Height, valuehash.Hash, base.Manifest) (bool, error)) error {
	var dir int = 1
	if reverse {
		dir = -1
	}

	opt := options.Find().
		SetSort(util.NewBSONFilter("height", dir).D()).
		SetLimit(limit)

	if !load {
		opt = opt.SetProjection(bson.M{"height": 1, "hash": 1})
	}

	return st.client.Find(
		context.Background(),
		ColNameManifest,
		filter,
		func(cursor *mongo.Cursor) (bool, error) {
			var height base.Height
			var h mitumutil.Hash
			var m base.Manifest

			if !load {
				if ht, i, err := loadManifestHeightAndHash(cursor.Decode, st.encs); err != nil {
					return false, err
				} else {
					height = ht
					h = i
				}
			} else {
				if i, err := loadManifestFromDecoder(cursor.Decode, st.encs); err != nil {
					return false, err
				} else {
					height = i.Height()
					h = i.Hash()
					m = i
				}
			}

			return callback(height, h, m)
		},
		opt,
	)
}

func (st *Database) NewOperationSeals(seals []operation.Seal) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if len(seals) < 1 {
		return errors.Errorf("empty seals")
	}

	var models []mongo.WriteModel

	filter := st.newStagedOperationFilter()

	for i := range seals {
		sl := seals[i]

		ms, err := st.newOperations(sl.Operations(), filter)
		if err != nil {
			return err
		}

		models = append(models, ms...)
	}

	return st.client.Bulk(context.Background(), ColNameStagedOperation, models, false)
}
*/

/*
func (st *Database) NewOperations(
	ops []base.Operation,
) error {
	filter := st.newStagedOperationFilter()

	models, err := st.newOperations(ops, filter)
	if err != nil {
		return err
	}

	if err := st.client.Bulk(context.Background(), ColNameStagedOperation, models, false); err != nil {
		return err
	}

	return nil
}

func (st *Database) newOperations(
	ops []base.Operation,
	filter func(valuehash.Hash) (bool, error),
) ([]mongo.WriteModel, error) {
	models := make([]mongo.WriteModel, len(ops))
	for i := range ops {
		op := ops[i]
		switch found, err := filter(op.Fact().Hash()); {
		case err != nil:
			return nil, err
		case !found:
			continue
		}

		doc, err := NewStagedOperation(op, st.enc)
		if err != nil {
			return nil, err
		}
		m, err := doc.bsonM()
		if err != nil {
			return nil, err
		}
		delete(m, "_id")

		models[i] = mongo.NewUpdateManyModel().
			SetFilter(util.NewBSONFilter("_id", op.Fact().Hash().String()).D()).
			SetUpdate(bson.D{{Key: "$set", Value: m}}).
			SetUpsert(true)
	}

	return models, nil
}

func (st *Database) StagedOperations(callback func(base.Operation) (bool, error), sort bool) error {
	var dir int
	if sort {
		dir = 1
	} else {
		dir = -1
	}

	opt := options.Find()
	opt.SetSort(util.NewBSONFilter("inserted_at", dir).D())

	return st.client.Find(
		context.TODO(),
		ColNameStagedOperation,
		bson.D{},
		func(cursor *mongo.Cursor) (bool, error) {
			op, err := loadStagedOperationFromDecoder(cursor.Decode, st.encs)
			if err != nil {
				return false, err
			}

			return callback(op)
		},
		opt,
	)
}

func (st *Database) UnstagedOperations(facts []valuehash.Hash) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	var models []mongo.WriteModel
	for i := range facts {
		models = append(models,
			mongo.NewDeleteOneModel().SetFilter(util.NewBSONFilter("_id", facts[i].String()).D()),
		)
	}

	return st.client.Bulk(context.Background(), ColNameStagedOperation, models, false)
}

func (st *Database) StagedOperationsByFact(facts []valuehash.Hash) ([]base.Operation, error) {
	var ops []base.Operation
	for i := range facts {
		h := facts[i]

		if err := st.client.GetByFilter(
			ColNameStagedOperation,
			util.NewBSONFilter("_id", h.String()).D(),
			func(res *mongo.SingleResult) error {
				op, err := loadStagedOperationFromDecoder(res.Decode, st.encs)
				if err != nil {
					return err
				}

				ops = append(ops, op)

				return nil
			},
		); err != nil {
			if errors.Is(err, mitumutil.ErrNotFound) {
				continue
			}

			return nil, err
		}
	}

	return ops, nil
}

func (st *Database) HasStagedOperation(h mitumutil.Hash) (bool, error) {
	count, err := st.client.Count(
		context.Background(),
		ColNameStagedOperation,
		util.NewBSONFilter("_id", h.String()).D(),
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (st *Database) State(key string) (base.State, bool, error) {
	if i, _ := st.stateCache.Get(key); i != nil {
		return i.(base.State), true, nil
	}

	var sta base.State

	if err := st.client.Find(
		context.TODO(),
		ColNameState,
		util.NewBSONFilter("key", key).AddOp("height", st.lastHeight(), "$lte").D(),
		func(cursor *mongo.Cursor) (bool, error) {
			if i, err := loadStateFromDecoder(cursor.Decode, st.encs); err != nil {
				return false, err
			} else {
				sta = i
			}

			return false, nil
		},
		options.Find().SetSort(util.NewBSONFilter("height", -1).D()).SetLimit(1),
	); err != nil {
		return nil, false, err
	}

	return sta, sta != nil, nil
}

func (st *Database) NewState(sta base.State) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if doc, err := NewStateDoc(sta, st.enc); err != nil {
		return err
	} else if _, err := st.client.Add(ColNameState, doc); err != nil {
		return err
	}

	_ = st.stateCache.Set(sta.Key(), sta, 0)

	return nil
}

func (st *Database) HasOperationFact(h valuehash.Hash) (bool, error) {
	if st.operationFactCache.Has(h.String()) {
		return true, nil
	}

	count, err := st.client.Count(
		context.Background(),
		ColNameOperation,
		util.NewBSONFilter("fact", h.String()).AddOp("height", st.lastHeight(), "$lte").D(),
		options.Count().SetLimit(1),
	)
	if err != nil {
		return false, err
	}

	if count > 0 {
		_ = st.operationFactCache.Set(h.String(), struct{}{}, 0)
	}

	return count > 0, nil
}
*/

/*

func (st *Database) NewSession(blk base.Block) (storage.DatabaseSession, error) {
	if st.readonly {
		return nil, errors.Errorf("readonly mode")
	}

	return NewDatabaseSession(st, blk)
}

*/

func (st *Database) initialize() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	for col, models := range defaultIndexes {
		if err := st.CreateIndex(col, models, IndexPrefix); err != nil {
			return err
		}
	}

	return nil
}

// Clean will drop the existing collections. To keep safe the another
// collections by user, drop collections instead of drop database.
func (st *Database) Clean() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	drop := func(c string) error {
		return st.client.Collection(c).Drop(context.Background())
	}

	for _, c := range allCollections {
		if err := drop(c); err != nil {
			return err
		}
	}

	if err := st.initialize(); err != nil {
		return err
	}

	st.Lock()
	defer st.Unlock()

	st.lastManifest = nil
	st.lastManifestHeight = base.NilHeight
	st.lastINITVoteproof = nil
	st.lastACCEPTVoteproof = nil

	return nil
}

func (st *Database) cleanByHeight(height base.Height) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if height <= base.GenesisHeight {
		return st.Clean()
	}

	opts := options.BulkWrite().SetOrdered(true)
	removeByHeight := mongo.NewDeleteManyModel().SetFilter(bson.M{"height": bson.M{"$gte": height}})

	for _, col := range allCollections {
		res, err := st.client.Collection(col).BulkWrite(
			context.Background(),
			[]mongo.WriteModel{removeByHeight},
			opts,
		)
		if err != nil {
			return err
		}

		st.Log().Debug().Str("collection", col).Interface("result", res).Msg("clean collection by height")
	}

	return nil
}

/*

func (st *Database) cleanupIncompleteData() error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	return st.cleanByHeight(st.lastHeight() + 1)
}

*/

func (st *Database) CreateIndex(col string, models []mongo.IndexModel, prefix string) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	st.Lock()
	defer st.Unlock()

	iv := st.client.Collection(col).Indexes()

	cursor, err := iv.List(context.TODO())
	if err != nil {
		return err
	}

	var existings []string
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return err
	} else {
		for _, r := range results {
			name := r["name"].(string)
			if !isIndexName(name, prefix) {
				continue
			}

			existings = append(existings, name)
		}
	}

	if len(existings) > 0 {
		for _, name := range existings {
			if _, err := iv.DropOne(context.TODO(), name); err != nil {
				return err
			}
		}
	}

	if len(models) < 1 {
		return nil
	}

	if _, err := iv.CreateMany(context.TODO(), models); err != nil {
		return err
	}

	return nil
}

func (st *Database) New() (*Database, error) {
	var client *Client
	if cl, err := st.client.New(""); err != nil {
		return nil, err
	} else {
		client = cl
	}

	st.RLock()
	defer st.RUnlock()

	if nst, err := NewDatabase(client, st.encs, st.enc); err != nil {
		return nil, err
	} else {
		nst.lastManifest = st.lastManifest
		nst.lastManifestHeight = st.lastManifestHeight

		return nst, nil
	}
}

func (st *Database) SetInfo(key string, b []byte) error {
	if st.readonly {
		return errors.Errorf("readonly mode")
	}

	if doc, err := NewInfoDoc(key, b, st.enc); err != nil {
		return err
	} else if _, err := st.client.Set(ColNameInfo, doc); err != nil {
		return err
	} else {
		return nil
	}
}

func (st *Database) Info(key string) ([]byte, bool, error) {
	var b []byte
	if err := st.client.GetByID(ColNameInfo, infoDocKey(key),
		func(res *mongo.SingleResult) error {
			if i, err := loadInfo(res.Decode, st.encs); err != nil {
				return err
			} else {
				b = i
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, mitumutil.ErrNotFound) || errors.Is(err, mongo.ErrNoDocuments) {
			return nil, false, nil
		}

		return nil, false, err
	}

	return b, b != nil, nil
}

func (st *Database) Readonly() (*Database, error) {
	if nst, err := st.New(); err != nil {
		return nil, err
	} else {
		nst.readonly = true

		return nst, nil
	}
}

/*
func (st *Database) voteproofByFilter(filter bson.D) (base.Voteproof, bool, error) {
	var voteproof base.Voteproof

	if err := st.client.GetByFilter(
		ColNameVoteproof,
		filter,
		func(res *mongo.SingleResult) error {
			if i, err := loadVoteproofFromDecoder(res.Decode, st.encs); err != nil {
				return err
			} else {
				voteproof = i
			}

			return nil
		},
		options.FindOne().SetSort(util.NewBSONFilter("height", -1).D()),
	); err != nil {
		if errors.Is(err, mitumutil.ErrNotFound) {
			return nil, false, nil
		}

		return nil, false, err
	}

	if voteproof == nil {
		return nil, false, nil
	}

	return voteproof, true, nil
}

func (st *Database) lastVoteproofs(height base.Height) (base.Voteproof, base.Voteproof, error) {
	var initVoteproof, acceptVoteproof base.Voteproof
	switch i, found, err := st.voteproofByFilter(util.NewBSONFilter("height", height).Add("stage", base.StageINIT.String()).D()); {
	case err != nil:
		return nil, nil, errors.Wrapf(err, "find last init voteproof of height, %v", height)
	case !found:
		return nil, nil, mitumutil.ErrNotFound.Errorf("find last init voteproof of height, %v", height)
	default:
		initVoteproof = i
	}

	switch i, found, err := st.voteproofByFilter(util.NewBSONFilter("height", height).Add("stage", base.StageACCEPT.String()).D()); {
	case err != nil:
		return nil, nil, errors.Wrapf(err, "find last accept voteproof of height, %v", height)
	case !found:
		return nil, nil, mitumutil.ErrNotFound.Errorf("find last accept voteproof of height, %v", height)
	default:
		acceptVoteproof = i
	}

	return initVoteproof, acceptVoteproof, nil
}

func (st *Database) newStagedOperationFilter() func(mitumutil.Hash) (bool, error) {
	inserted := map[string]struct{}{}
	return func(h mitumutil.Hash) (bool, error) {
		k := h.String()
		if _, found := inserted[k]; found {
			return false, nil
		}

		switch found, err := st.HasOperationFact(h); {
		case err != nil:
			return false, err
		case found:
			return false, nil
		}

		switch found, err := st.HasStagedOperation(h); {
		case err != nil:
			return false, err
		case found:
			return false, nil
		}

		inserted[k] = struct{}{}

		return true, nil
	}
}
*/
