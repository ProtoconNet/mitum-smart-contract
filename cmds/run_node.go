package cmds

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	capi "github.com/imfact-labs/currency-model/api"
	ccmds "github.com/imfact-labs/currency-model/app/cmds"
	cpipeline "github.com/imfact-labs/currency-model/app/runtime/pipeline"
	cdigest "github.com/imfact-labs/currency-model/digest"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/isaac"
	isaacstates "github.com/imfact-labs/mitum2/isaac/states"
	"github.com/imfact-labs/mitum2/launch"
	"github.com/imfact-labs/mitum2/network/quicstream"
	"github.com/imfact-labs/mitum2/util"
	"github.com/imfact-labs/mitum2/util/logging"
	"github.com/imfact-labs/mitum2/util/ps"
	"github.com/imfact-labs/smart-contract-model/digest"
	"github.com/imfact-labs/smart-contract-model/runtime/steps"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type RunCommand struct { //nolint:govet //...
	ccmds.RunCommand

	//revive:disable:line-length-limit
	exitf  func(error)
	log    *zerolog.Logger
	holded bool
	//revive:enable:line-length-limit
}

func (cmd *RunCommand) Run(pctx context.Context) error {
	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return err
	}

	log.Log().Debug().
		Interface("design", cmd.DesignFlag).
		Interface("privatekey", cmd.PrivatekeyFlags).
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold).
		Interface("http_state", cmd.HTTPState).
		Interface("dev", cmd.DevFlags).
		Interface("acl", cmd.ACLFlags).
		Msg("flags")

	cmd.log = log.Log()
	cmd.RunCommand.SetLog(log.Log())

	if len(cmd.HTTPState) > 0 {
		if err := cmd.RunCommand.RunHTTPState(cmd.HTTPState); err != nil {
			return errors.Wrap(err, "failed to run http state")
		}
	}

	nctx := util.ContextWithValues(pctx, map[util.ContextKey]interface{}{
		launch.DesignFlagContextKey:    cmd.DesignFlag,
		launch.DevFlagsContextKey:      cmd.DevFlags,
		launch.DiscoveryFlagContextKey: cmd.Discovery,
		launch.PrivatekeyContextKey:    string(cmd.PrivatekeyFlags.Flag.Body()),
		launch.ACLFlagsContextKey:      cmd.ACLFlags,
	})

	pps := cpipeline.DefaultRunPS()
	_ = pps.POK(launch.PNameStates).PreRemoveOK(launch.PNameProposalProcessors)
	_ = pps.AddOK(PNameDigester, ProcessDigester, nil, cdigest.PNameDigesterDataBase).
		AddOK(PNameStartDigester, ProcessStartDigester, nil, capi.PNameStartAPI)
	_ = pps.POK(launch.PNameStorage).PostAddOK(ps.Name("check-hold"), cmd.pCheckHold)
	_ = pps.POK(launch.PNameStates).
		PreAddOK(steps.PNameOperationProcessorsMap, steps.POperationProcessorsMap).
		PreAddOK(launch.PNameProposalProcessors, PProposalProcessors).
		PreAddOK(ps.Name("when-new-block-saved-in-consensus-state-func"), cmd.pWhenNewBlockSavedInConsensusStateFunc).
		PreAddOK(ps.Name("when-new-block-confirmed-func"), cmd.pWhenNewBlockConfirmed).
		PreAddOK(ps.Name("when-new-block-saved-in-syncing-state-func"), cmd.pWhenNewBlockSavedInSyncingStateFunc)
	_ = pps.POK(launch.PNameEncoder).PostAddOK(launch.PNameAddHinters, steps.PAddHinters)
	_ = pps.POK(capi.PNameAPI).PostAddOK(ccmds.PNameDigestAPIHandlers, cmd.pDigestAPIHandlers)
	_ = pps.POK(PNameDigester).PostAddOK(ccmds.PNameDigesterFollowUp, PdigesterFollowUp)

	_ = pps.SetLogging(log)

	log.Log().Debug().Interface("process", pps.Verbose()).Msg("process ready")

	nctx, err := pps.Run(nctx)
	defer func() {
		log.Log().Debug().Interface("process", pps.Verbose()).Msg("process will be closed")

		if _, err = pps.Close(nctx); err != nil {
			log.Log().Error().Err(err).Msg("failed to close")
		}
	}()

	if err != nil {
		return err
	}

	log.Log().Debug().
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold.Height()).
		Msg("node started")

	return cmd.run(nctx)
}

var errHoldStop = util.NewIDError("hold stop")

func (cmd *RunCommand) run(pctx context.Context) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	exitch := make(chan error)
	cmd.exitf = func(err error) { exitch <- err }

	stopstates := func() {}
	if !cmd.holded {
		deferred, err := cmd.runStates(ctx, pctx)
		if err != nil {
			return err
		}
		stopstates = deferred
	}

	select {
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	case err := <-exitch:
		if errors.Is(err, errHoldStop) {
			stopstates()
			<-ctx.Done()
			return errors.WithStack(ctx.Err())
		}
		return err
	}
}

func (cmd *RunCommand) runStates(ctx, pctx context.Context) (func(), error) {
	var discoveries *util.Locked[[]quicstream.ConnInfo]
	var states *isaacstates.States
	if err := util.LoadFromContextOK(pctx,
		launch.DiscoveryContextKey, &discoveries,
		launch.StatesContextKey, &states,
	); err != nil {
		return nil, err
	}
	if dis := launch.GetDiscoveriesFromLocked(discoveries); len(dis) < 1 {
		cmd.log.Warn().Msg("empty discoveries; will wait to be joined by remote nodes")
	}

	go func() { cmd.exitf(<-states.Wait(ctx)) }()
	return func() {
		if err := states.Hold(); err != nil && !errors.Is(err, util.ErrDaemonAlreadyStopped) {
			cmd.log.Error().Err(err).Msg("stop states")
			return
		}
		cmd.log.Debug().Msg("states stopped")
	}, nil
}

func (cmd *RunCommand) pWhenNewBlockSavedInSyncingStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var di *digest.Digester

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return pctx, err
	}

	if err := util.LoadFromContext(pctx, cdigest.ContextValueDigester, &di); err != nil {
		return pctx, err
	}

	var f func(height base.Height)
	if di != nil {
		g := cmd.whenBlockSaved(db, di)

		f = func(height base.Height) {
			g(pctx)
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	} else {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	}

	return context.WithValue(pctx,
		launch.WhenNewBlockSavedInSyncingStateFuncContextKey, f,
	), nil
}

func (cmd *RunCommand) pWhenNewBlockSavedInConsensusStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
	); err != nil {
		return pctx, err
	}

	f := func(bm base.BlockMap) {
		l := log.Log().With().
			Interface("blockmap", bm).
			Interface("height", bm.Manifest().Height()).
			Logger()

		if cmd.Hold.IsSet() && bm.Manifest().Height() == cmd.Hold.Height() {
			l.Debug().Msg("will be stopped by hold")

			cmd.exitf(errHoldStop.WithStack())

			return
		}
	}

	return context.WithValue(pctx, launch.WhenNewBlockSavedInConsensusStateFuncContextKey, f), nil
}

func (cmd *RunCommand) pWhenNewBlockConfirmed(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var di *digest.Digester

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
	); err != nil {
		return pctx, err
	}

	if err := util.LoadFromContext(pctx, cdigest.ContextValueDigester, &di); err != nil {
		return pctx, err
	}

	var f func(height base.Height)
	if di != nil {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()
			err := digestFollowup(pctx, height)
			if err != nil {
				cmd.exitf(err)

				return
			}

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	} else {
		f = func(height base.Height) {
			l := log.Log().With().Interface("height", height).Logger()

			if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
				l.Debug().Msg("will be stopped by hold")
				cmd.exitf(errHoldStop.WithStack())

				return
			}
		}
	}

	return context.WithValue(pctx,
		launch.WhenNewBlockConfirmedFuncContextKey, f,
	), nil
}

func (cmd *RunCommand) whenBlockSaved(
	db isaac.Database,
	di *digest.Digester,
) ps.Func {
	return func(ctx context.Context) (context.Context, error) {
		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return ctx, err
		case !found:
			return ctx, errors.Errorf("Last BlockMap not found")
		default:
			if di != nil {
				go func() {
					di.Digest([]base.BlockMap{m})
				}()
			}
		}
		return ctx, nil
	}
}

func (cmd *RunCommand) pCheckHold(pctx context.Context) (context.Context, error) {
	var db isaac.Database
	if err := util.LoadFromContextOK(pctx, launch.CenterDatabaseContextKey, &db); err != nil {
		return pctx, err
	}

	switch {
	case !cmd.Hold.IsSet():
	case cmd.Hold.Height() < base.GenesisHeight:
		cmd.holded = true
	default:
		switch m, found, err := db.LastBlockMap(); {
		case err != nil:
			return pctx, err
		case !found:
		case cmd.Hold.Height() <= m.Manifest().Height():
			cmd.holded = true
		}
	}

	return pctx, nil
}

func (cmd *RunCommand) pDigestAPIHandlers(ctx context.Context) (context.Context, error) {
	var params *launch.LocalParams
	var local base.LocalNode

	if err := util.LoadFromContextOK(ctx,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
	); err != nil {
		return nil, err
	}

	var design cdigest.YamlDigestDesign
	if err := util.LoadFromContext(ctx, cdigest.ContextValueDigestDesign, &design); err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return ctx, nil
		}

		return nil, err
	}

	if design.Equal(cdigest.YamlDigestDesign{}) {
		return ctx, nil
	}

	cache, err := ccmds.LoadCache(cmd.RunCommand.Log(), ctx, design)
	if err != nil {
		return ctx, err
	}

	var dnt *capi.HTTP2Server
	if err := util.LoadFromContext(ctx, capi.ContextValueDigestNetwork, &dnt); err != nil {
		return ctx, err
	}
	router := dnt.Router()

	defaultHandlers, err := ccmds.SetDigestAPIDefaultHandlers(
		cmd.RunCommand.Log(), ctx, params, cache, router, dnt.Queue(),
	)
	if err != nil {
		return ctx, err
	}

	if err := defaultHandlers.Initialize(); err != nil {
		return ctx, err
	}
	capi.SetHandlers(defaultHandlers, design.Digest)
	defaultHandlers.SetEncoders(encs)
	defaultHandlers.SetEncoder(enc)

	handlers, err := cmd.setDigestHandlers(ctx, params, cache, router, defaultHandlers.Routes())
	if err != nil {
		return ctx, err
	}

	if err := handlers.Initialize(); err != nil {
		return ctx, err
	}

	dnt.SetEncoder(encs)

	return ctx, nil
}

func (cmd *RunCommand) setDigestHandlers(
	ctx context.Context,
	params *launch.LocalParams,
	cache capi.Cache,
	router *mux.Router,
	routes map[string]*mux.Route,
) (*digest.Handlers, error) {
	var st *cdigest.Database
	if err := util.LoadFromContext(ctx, cdigest.ContextValueDigestDatabase, &st); err != nil {
		return nil, err
	}

	handlers := digest.NewHandlers(ctx, params.ISAAC.NetworkID(), encs, enc, st, cache, router, routes)

	return handlers, nil
}
