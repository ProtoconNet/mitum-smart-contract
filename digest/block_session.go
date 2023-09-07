package digest

import (
	"context"
	"fmt"
	statecurrency "github.com/ProtoconNet/mitum-currency/v3/state/currency"
	stateextension "github.com/ProtoconNet/mitum-currency/v3/state/extension"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/ProtoconNet/mitum-currency/v3/digest/isaac"
	"github.com/ProtoconNet/mitum2/base"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/fixedtree"
)

var bulkWriteLimit = 500

type BlockSessioner interface {
	Prepare() error
	Commit(context.Context) error
	Close() error
}

type BlockSession struct {
	sync.RWMutex
	block                 base.BlockMap
	ops                   []base.Operation
	opsTree               fixedtree.Tree
	sts                   []base.State
	st                    *Database
	opsTreeNodes          map[string]base.OperationFixedtreeNode
	blockModels           []mongo.WriteModel
	operationModels       []mongo.WriteModel
	accountModels         []mongo.WriteModel
	contractAccountModels []mongo.WriteModel
	balanceModels         []mongo.WriteModel
	currencyModels        []mongo.WriteModel
	statesValue           *sync.Map
	balanceAddressList    []string
}

func NewBlockSession(st *Database, blk base.BlockMap, ops []base.Operation, opsTree fixedtree.Tree, sts []base.State) (*BlockSession, error) {
	if st.Readonly() {
		return nil, errors.Errorf("readonly mode")
	}

	nst, err := st.New()
	if err != nil {
		return nil, err
	}

	return &BlockSession{
		st:          nst,
		block:       blk,
		ops:         ops,
		opsTree:     opsTree,
		sts:         sts,
		statesValue: &sync.Map{},
	}, nil
}

func (bs *BlockSession) Prepare() error {
	bs.Lock()
	defer bs.Unlock()

	if err := bs.prepareOperationsTree(); err != nil {
		return err
	}

	if err := bs.prepareBlock(); err != nil {
		return err
	}

	if err := bs.prepareOperations(); err != nil {
		return err
	}

	if err := bs.prepareCurrencies(); err != nil {
		return err
	}

	return bs.prepareAccounts()
}

func (bs *BlockSession) Commit(ctx context.Context) error {
	bs.Lock()
	defer bs.Unlock()

	started := time.Now()
	defer func() {
		bs.statesValue.Store("commit", time.Since(started))

		_ = bs.close()
	}()

	if err := bs.writeModels(ctx, defaultColNameBlock, bs.blockModels); err != nil {
		return err
	}

	if len(bs.operationModels) > 0 {

	}
	if err := bs.writeModels(ctx, defaultColNameOperation, bs.operationModels); err != nil {
		return err
	}

	if len(bs.currencyModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameCurrency, bs.currencyModels); err != nil {
			return err
		}
	}

	if len(bs.accountModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameAccount, bs.accountModels); err != nil {
			return err
		}
	}

	if len(bs.contractAccountModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameContractAccount, bs.contractAccountModels); err != nil {
			return err
		}
	}

	if len(bs.balanceModels) > 0 {
		if err := bs.writeModels(ctx, defaultColNameBalance, bs.balanceModels); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BlockSession) Close() error {
	bs.Lock()
	defer bs.Unlock()

	return bs.close()
}

func (bs *BlockSession) prepareOperationsTree() error {
	nodes := map[string]base.OperationFixedtreeNode{}

	if err := bs.opsTree.Traverse(func(_ uint64, no fixedtree.Node) (bool, error) {
		nno := no.(base.OperationFixedtreeNode)
		nodes[nno.Key()] = nno

		return true, nil
	}); err != nil {
		return err
	}

	bs.opsTreeNodes = nodes

	return nil
}

func (bs *BlockSession) prepareBlock() error {
	if bs.block == nil {
		return nil
	}

	bs.blockModels = make([]mongo.WriteModel, 1)

	manifest := isaac.NewManifest(
		bs.block.Manifest().Height(),
		bs.block.Manifest().Previous(),
		bs.block.Manifest().Proposal(),
		bs.block.Manifest().OperationsTree(),
		bs.block.Manifest().StatesTree(),
		bs.block.Manifest().Suffrage(),
		bs.block.Manifest().ProposedAt(),
	)

	doc, err := NewManifestDoc(manifest, bs.st.database.Encoder(), bs.block.Manifest().Height(), bs.ops, bs.block.SignedAt())
	if err != nil {
		return err
	}
	bs.blockModels[0] = mongo.NewInsertOneModel().SetDocument(doc)

	return nil
}

func (bs *BlockSession) prepareOperations() error {
	if len(bs.ops) < 1 {
		return nil
	}

	node := func(h mitumutil.Hash) (bool, bool, base.OperationProcessReasonError) {
		no, found := bs.opsTreeNodes[h.String()]
		if !found {
			return false, false, nil
		}

		return true, no.InState(), no.Reason()
	}

	bs.operationModels = make([]mongo.WriteModel, len(bs.ops))

	for i := range bs.ops {
		op := bs.ops[i]

		found, inState, reason := node(op.Fact().Hash())
		if !found {
			return mitumutil.ErrNotFound.Errorf("operation, %v in operations tree", op.Fact().Hash().String())
		}

		doc, err := NewOperationDoc(
			op,
			bs.st.database.Encoder(),
			bs.block.Manifest().Height(),
			bs.block.SignedAt(),
			inState,
			reason,
			uint64(i),
		)
		if err != nil {
			return err
		}
		bs.operationModels[i] = mongo.NewInsertOneModel().SetDocument(doc)
	}

	return nil
}

func (bs *BlockSession) prepareAccounts() error {
	if len(bs.sts) < 1 {
		return nil
	}

	var accountModels []mongo.WriteModel
	var balanceModels []mongo.WriteModel
	var contractAccountModels []mongo.WriteModel
	for i := range bs.sts {
		st := bs.sts[i]

		switch {
		case statecurrency.IsStateAccountKey(st.Key()):
			j, err := bs.handleAccountState(st)
			if err != nil {
				return err
			}
			accountModels = append(accountModels, j...)
		case statecurrency.IsStateBalanceKey(st.Key()):
			j, address, err := bs.handleBalanceState(st)
			if err != nil {
				return err
			}
			balanceModels = append(balanceModels, j...)
			bs.balanceAddressList = append(bs.balanceAddressList, address)
		case stateextension.IsStateContractAccountKey(st.Key()):
			j, err := bs.handleContractAccountState(st)
			if err != nil {
				return err
			}
			contractAccountModels = append(contractAccountModels, j...)
		default:
			continue
		}
	}

	bs.accountModels = accountModels
	bs.contractAccountModels = contractAccountModels
	bs.balanceModels = balanceModels
	return nil
}

func (bs *BlockSession) prepareCurrencies() error {
	if len(bs.sts) < 1 {
		return nil
	}

	var currencyModels []mongo.WriteModel
	for i := range bs.sts {
		st := bs.sts[i]
		switch {
		case statecurrency.IsStateCurrencyDesignKey(st.Key()):
			j, err := bs.handleCurrencyState(st)
			if err != nil {
				return err
			}
			currencyModels = append(currencyModels, j...)
		default:
			continue
		}
	}

	bs.currencyModels = currencyModels

	return nil
}

func (bs *BlockSession) handleAccountState(st base.State) ([]mongo.WriteModel, error) {
	if rs, err := NewAccountValue(st); err != nil {
		return nil, err
	} else if doc, err := NewAccountDoc(rs, bs.st.database.Encoder()); err != nil {
		return nil, err
	} else {
		return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
	}
}

func (bs *BlockSession) handleBalanceState(st base.State) ([]mongo.WriteModel, string, error) {
	doc, address, err := NewBalanceDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, "", err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, address, nil
}

func (bs *BlockSession) handleContractAccountState(st base.State) ([]mongo.WriteModel, error) {
	doc, err := NewContractAccountStatusDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) handleCurrencyState(st base.State) ([]mongo.WriteModel, error) {
	doc, err := NewCurrencyDoc(st, bs.st.database.Encoder())
	if err != nil {
		return nil, err
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(doc)}, nil
}

func (bs *BlockSession) writeModels(ctx context.Context, col string, models []mongo.WriteModel) error {
	started := time.Now()
	defer func() {
		bs.statesValue.Store(fmt.Sprintf("write-models-%s", col), time.Since(started))
	}()

	n := len(models)
	if n < 1 {
		return nil
	} else if n <= bulkWriteLimit {
		return bs.writeModelsChunk(ctx, col, models)
	}

	z := n / bulkWriteLimit
	if n%bulkWriteLimit != 0 {
		z++
	}

	for i := 0; i < z; i++ {
		s := i * bulkWriteLimit
		e := s + bulkWriteLimit
		if e > n {
			e = n
		}

		if err := bs.writeModelsChunk(ctx, col, models[s:e]); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BlockSession) writeModelsChunk(ctx context.Context, col string, models []mongo.WriteModel) error {
	opts := options.BulkWrite().SetOrdered(false)
	if res, err := bs.st.database.Client().Collection(col).BulkWrite(ctx, models, opts); err != nil {
		return err
	} else if res != nil && res.InsertedCount < 1 {
		return errors.Errorf("not inserted to %s", col)
	}

	return nil
}

func (bs *BlockSession) close() error {
	bs.block = nil
	bs.operationModels = nil
	bs.currencyModels = nil
	bs.accountModels = nil
	bs.contractAccountModels = nil
	bs.balanceModels = nil

	return bs.st.Close()
}
