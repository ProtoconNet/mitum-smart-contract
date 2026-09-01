package cmds

import (
	"context"
	"math"
	"time"

	"github.com/imfact-labs/mitum2/launch"

	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	isaacdatabase "github.com/imfact-labs/mitum2/isaac/database"
	"github.com/imfact-labs/mitum2/network/quicmemberlist"
	"github.com/imfact-labs/mitum2/network/quicstream"
	"github.com/imfact-labs/mitum2/storage"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/mitum2/util/hint"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/smart-contract-model/operation/processor"
	"github.com/imfact-labs/smart-contract-model/runtime/contracts"
	"github.com/pkg/errors"
)

func PProposalProcessors(pctx context.Context) (context.Context, error) {
	var log *logging.Logging

	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return pctx, err
	}

	newProposalProcessorf, err := newProposalProcessorFunc(pctx)
	if err != nil {
		return pctx, err
	}

	getProposalf, err := getProposalFunc(pctx)
	if err != nil {
		return pctx, err
	}

	pps := isaac.NewProposalProcessors(newProposalProcessorf, getProposalf)
	_ = pps.SetLogging(log)

	return context.WithValue(pctx, launch.ProposalProcessorsContextKey, pps), nil
}

func newProposalProcessorFunc(pctx context.Context) (
	func(base.ProposalSignFact, base.Manifest) (isaac.ProposalProcessor, error),
	error,
) {
	var encs *encoder.Encoders
	var design launch.NodeDesign
	var local base.LocalNode
	var isaacparams *isaac.Params
	var db isaac.Database
	var oprs *hint.CompatibleSet[isaac.NewOperationProcessorInternalFunc]
	var oprsB *hint.CompatibleSet[contracts.NewOperationProcessorInternalWithProposalFunc]

	if err := util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &encs,
		launch.DesignContextKey, &design,
		launch.LocalContextKey, &local,
		launch.ISAACParamsContextKey, &isaacparams,
		launch.CenterDatabaseContextKey, &db,
		launch.OperationProcessorsMapContextKey, &oprs,
		contracts.OperationProcessorsMapBContextKey, &oprsB,
	); err != nil {
		return nil, err
	}

	getProposalOperationFuncf, err := getProposalOperationFunc(pctx)
	if err != nil {
		return nil, err
	}

	return func(proposal base.ProposalSignFact, previous base.Manifest) (
		isaac.ProposalProcessor, error,
	) {
		args := isaac.NewDefaultProposalProcessorArgs()
		args.MaxWorkerSize = math.MaxInt16
		args.NewWriterFunc = launch.NewBlockWriterFunc(
			local,
			isaacparams.NetworkID(),
			launch.LocalFSDataDirectory(design.Storage.Base),
			encs.JSON(),
			encs.Default(),
			db,
			args.MaxWorkerSize,
			isaacparams.StateCacheSize(),
		)
		args.GetStateFunc = db.State
		args.GetOperationFunc = getProposalOperationFuncf(proposal)
		args.OperationExecutionClassFunc = processor.ExecutionClass
		args.NewOperationProcessorFunc = func(height base.Height, ht hint.Hint, getStatef base.GetStateFunc,
		) (base.OperationProcessor, error) {
			v, found := oprs.Find(ht)
			if found {
				return v(height, getStatef)
			}

			w, found := oprsB.Find(ht)
			if found {
				return w(height, proposal, getStatef)
			}

			return nil, nil
		}
		args.EmptyProposalNoBlockFunc = func() bool {
			return db.LastNetworkPolicy().EmptyProposalNoBlock()
		}

		return isaac.NewDefaultProposalProcessor(proposal, previous, args)
	}, nil
}

func getProposalFunc(pctx context.Context) (
	func(context.Context, base.Point, util.Hash) (base.ProposalSignFact, error),
	error,
) {
	var params *launch.LocalParams
	var pool *isaacdatabase.TempPool
	var client isaac.NetworkClient
	var m *quicmemberlist.Memberlist

	if err := util.LoadFromContextOK(pctx,
		launch.LocalParamsContextKey, &params,
		launch.PoolDatabaseContextKey, &pool,
		launch.QuicstreamClientContextKey, &client,
		launch.MemberlistContextKey, &m,
	); err != nil {
		return nil, err
	}

	return func(ctx context.Context, point base.Point, facthash util.Hash) (base.ProposalSignFact, error) {
		switch pr, found, err := pool.Proposal(facthash); {
		case err != nil:
			return nil, err
		case found:
			return pr, nil
		}

		semsize := int64(m.RemotesLen())
		if semsize < 1 {
			return nil, storage.ErrNotFound.Errorf("empty remote")
		}

		// NOTE if not found, request to remote node
		var worker *util.BaseJobWorker

		switch i, err := util.NewBaseJobWorker(ctx, semsize); {
		case err != nil:
			return nil, err
		default:
			worker = i
		}

		defer worker.Close()

		prl := util.EmptyLocked[base.ProposalSignFact]()

		go func() {
			defer worker.Done()

			m.Remotes(func(node quicmemberlist.Member) bool {
				ci := node.ConnInfo()

				return worker.NewJob(func(ctx context.Context, _ uint64) error {
					cctx, cancel := context.WithTimeout(ctx, params.Network.TimeoutRequest())
					defer cancel()

					var pr base.ProposalSignFact

					switch i, found, err := client.Proposal(cctx, ci, facthash); {
					case err != nil || !found:
						return nil
					default:
						if ierr := i.IsValid(params.ISAAC.NetworkID()); ierr != nil {
							return ierr
						}

						pr = i
					}

					switch {
					case !point.Equal(pr.Point()):
						return nil
					case !facthash.Equal(pr.Fact().Hash()):
						return nil
					}

					_ = prl.GetOrCreate(
						func(base.ProposalSignFact, bool) error {
							return nil
						},
						func() (base.ProposalSignFact, error) {
							return pr, nil
						},
					)

					return errors.Errorf("stop")
				}) == nil
			})
		}()

		err := worker.Wait()

		switch i, _ := prl.Value(); {
		case i == nil:
			if err != nil {
				return nil, err
			}

			return nil, storage.ErrNotFound.Errorf("ProposalSignFact not found")
		default:
			_, _ = pool.SetProposal(i)

			return i, nil
		}
	}, nil
}

func getProposalOperationFunc(pctx context.Context) (
	func(base.ProposalSignFact) isaac.OperationProcessorGetOperationFunction,
	error,
) {
	var isaacparams *isaac.Params
	var db isaac.Database

	if err := util.LoadFromContextOK(pctx,
		launch.ISAACParamsContextKey, &isaacparams,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return nil, err
	}

	getProposalOperationFromPoolf, err := getProposalOperationFromPoolFunc(pctx)
	if err != nil {
		return nil, err
	}

	getProposalOperationFromRemotef, err := getProposalOperationFromRemoteFunc(pctx)
	if err != nil {
		return nil, err
	}

	return func(proposal base.ProposalSignFact) isaac.OperationProcessorGetOperationFunction {
		return func(ctx context.Context, operationhash, fact util.Hash) (base.Operation, error) {
			switch found, err := db.ExistsInStateOperation(fact); {
			case err != nil:
				return nil, err
			case found:
				return nil, isaac.ErrOperationAlreadyProcessedInProcessor.Errorf("already processed")
			}

			switch i, found, err := getProposalOperationFromPoolf(ctx, operationhash); {
			case err != nil:
				return nil, err
			case found:
				return i, nil
			}

			var op base.Operation

			if err := util.Retry(
				ctx,
				func() (bool, error) {
					switch i, found, err := getProposalOperationFromRemotef(ctx, proposal, operationhash); {
					case err != nil:
						return true, isaac.ErrOperationNotFoundInProcessor.Wrap(err)
					case !found:
						return true, isaac.ErrOperationNotFoundInProcessor.Errorf("not found in remote")
					default:
						op = i

						return false, nil
					}
				},
				15,                   //nolint:gomnd //...
				time.Millisecond*333, //nolint:gomnd //...
			); err != nil {
				return nil, err
			}

			return op, nil
		}
	}, nil
}

func getProposalOperationFromPoolFunc(pctx context.Context) (
	func(pctx context.Context, operationhash util.Hash) (base.Operation, bool, error),
	error,
) {
	var pool *isaacdatabase.TempPool

	if err := util.LoadFromContextOK(pctx, launch.PoolDatabaseContextKey, &pool); err != nil {
		return nil, err
	}

	return func(ctx context.Context, operationhash util.Hash) (base.Operation, bool, error) {
		op, found, err := pool.Operation(ctx, operationhash)

		switch {
		case err != nil:
			return nil, false, err
		case !found:
			return nil, false, nil
		default:
			return op, true, nil
		}
	}, nil
}

func getProposalOperationFromRemoteFunc(pctx context.Context) ( //nolint:gocognit //...
	func(context.Context, base.ProposalSignFact, util.Hash) (base.Operation, bool, error),
	error,
) {
	var params *launch.LocalParams
	var client isaac.NetworkClient
	var syncSourcePool *isaac.SyncSourcePool

	if err := util.LoadFromContextOK(pctx,
		launch.LocalParamsContextKey, &params,
		launch.QuicstreamClientContextKey, &client,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
	); err != nil {
		return nil, err
	}

	getProposalOperationFromRemoteProposerf, err := getProposalOperationFromRemoteProposerFunc(pctx)
	if err != nil {
		return nil, err
	}

	return func(
		ctx context.Context, proposal base.ProposalSignFact, operationhash util.Hash,
	) (base.Operation, bool, error) {
		// NOTE try proposer first (incl. memberlist fallback) even when the sync source
		// pool is empty; the proposer holds its proposal's operations.
		var proposerErr error

		switch isproposer, op, found, err := getProposalOperationFromRemoteProposerf(ctx, proposal, operationhash); {
		case err != nil:
			if ctx.Err() != nil {
				return nil, false, errors.WithStack(ctx.Err())
			}

			proposerErr = err
		case !isproposer:
		case !found:
			// NOTE proposer proposed this operation, but it does not have? weired.
		default:
			return op, true, nil
		}

		proposer := proposal.ProposalFact().Proposer()

		if syncSourcePool.Len() < 1 {
			// NOTE proposer fetch (incl. memberlist fallback) already attempted above;
			// no sync source peers remain to ask.
			return nil, false, proposerErr
		}

		result := util.EmptyLocked[base.Operation]()

		worker, err := util.NewBaseJobWorker(ctx, int64(syncSourcePool.Len()))
		if err != nil {
			return nil, false, err
		}

		defer worker.Close()

		syncSourcePool.Actives(func(nci isaac.NodeConnInfo) bool {
			if proposer.Equal(nci.Address()) {
				return true
			}

			if werr := worker.NewJob(func(ctx context.Context, _ uint64) error {
				cctx, cancel := context.WithTimeout(ctx, params.Network.TimeoutRequest())
				defer cancel()

				op, _ := result.Set(func(i base.Operation, _ bool) (base.Operation, error) {
					if i != nil {
						return i, util.ErrLockedSetIgnore
					}

					switch op, found, jerr := client.Operation(cctx, nci.ConnInfo(), operationhash); {
					case jerr != nil:
						return nil, util.ErrLockedSetIgnore
					case !found:
						return nil, util.ErrLockedSetIgnore
					default:
						return op, util.ErrLockedSetIgnore
					}
				})

				if op != nil {
					return errors.Errorf("stop")
				}

				return nil
			}); werr != nil {
				return false
			}

			return true
		})

		worker.Done()

		err = worker.Wait()

		i, _ := result.Value()
		if i == nil {
			if ctx.Err() != nil {
				return nil, false, errors.WithStack(ctx.Err())
			}

			if proposerErr != nil {
				err = proposerErr
			}

			return nil, false, err
		}

		return i, true, nil
	}, nil
}

func getProposalOperationFromRemoteProposerFunc(pctx context.Context) (
	func(context.Context, base.ProposalSignFact, util.Hash) (bool, base.Operation, bool, error),
	error,
) {
	var params *launch.LocalParams
	var client isaac.NetworkClient
	var syncSourcePool *isaac.SyncSourcePool
	var m *quicmemberlist.Memberlist

	if err := util.LoadFromContextOK(pctx,
		launch.LocalParamsContextKey, &params,
		launch.QuicstreamClientContextKey, &client,
		launch.SyncSourcePoolContextKey, &syncSourcePool,
		launch.MemberlistContextKey, &m,
	); err != nil {
		return nil, err
	}

	return func(
		ctx context.Context, proposal base.ProposalSignFact, operationhash util.Hash,
	) (bool, base.Operation, bool, error) {
		proposer := proposal.ProposalFact().Proposer()

		// NOTE find proposer conn info from sync sources first, then from memberlist.
		// The proposer is the guaranteed holder of its proposal's operations, so it must
		// be reachable even when the sync source pool does not contain it (e.g. empty pool).
		var proposerci quicstream.ConnInfo

		var hasproposerci bool

		syncSourcePool.Actives(func(nci isaac.NodeConnInfo) bool {
			if !proposer.Equal(nci.Address()) {
				return true
			}

			proposerci = nci.ConnInfo()
			hasproposerci = true

			return false
		})

		if !hasproposerci {
			m.Members(func(member quicmemberlist.Member) bool {
				if !proposer.Equal(member.Address()) {
					return true
				}

				proposerci = member.ConnInfo()
				hasproposerci = true

				return false
			})
		}

		if !hasproposerci {
			return false, nil, false, nil
		}

		cctx, cancel := context.WithTimeout(ctx, params.Network.TimeoutRequest())
		defer cancel()

		switch op, found, err := client.Operation(cctx, proposerci, operationhash); {
		case err != nil:
			return true, nil, false, err
		case !found:
			return true, nil, false, nil
		default:
			return true, op, true, nil
		}
	}, nil
}
