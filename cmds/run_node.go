package cmds

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arl/statsviz"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum-currency/digest"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/isaac"
	isaacnetwork "github.com/spikeekips/mitum/isaac/network"
	isaacstates "github.com/spikeekips/mitum/isaac/states"
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/network/quicmemberlist"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
	"github.com/spikeekips/mitum/util/ps"
)

type RunCommand struct { //nolint:govet //...
	launch.DesignFlag
	launch.DevFlags `embed:"" prefix:"dev."`
	Vault           string                `name:"vault" help:"privatekey path of vault"`
	Discovery       []launch.ConnInfoFlag `help:"member discovery" placeholder:"ConnInfo"`
	Hold            launch.HeightFlag     `help:"hold consensus states"`
	HTTPState       string                `name:"http-state" help:"runtime statistics thru https" placeholder:"bind address"`
	exitf           func(error)
	log             *zerolog.Logger
	holded          bool
}

func NewRunCommand() RunCommand {
	return RunCommand{}
}

func (cmd *RunCommand) Run(pctx context.Context) error {
	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return err
	}

	log.Log().Debug().
		Interface("design", cmd.DesignFlag).
		Interface("vault", cmd.Vault).
		Interface("discovery", cmd.Discovery).
		Interface("hold", cmd.Hold).
		Interface("http_state", cmd.HTTPState).
		Interface("dev", cmd.DevFlags).
		Msg("flags")

	cmd.log = log.Log()

	if len(cmd.HTTPState) > 0 {
		if err := cmd.runHTTPState(cmd.HTTPState); err != nil {
			return errors.Wrap(err, "failed to run http state")
		}
	}

	//revive:disable:modifies-parameter
	pctx = context.WithValue(pctx, launch.DesignFlagContextKey, cmd.DesignFlag)
	pctx = context.WithValue(pctx, launch.DevFlagsContextKey, cmd.DevFlags)
	pctx = context.WithValue(pctx, launch.DiscoveryFlagContextKey, cmd.Discovery)
	pctx = context.WithValue(pctx, launch.VaultContextKey, cmd.Vault)
	//revive:enable:modifies-parameter

	pps := DefaultRunPS()

	_ = pps.POK(launch.PNameStorage).PostAddOK(ps.Name("check-hold"), cmd.pCheckHold)
	_ = pps.POK(launch.PNameStates).PreAddOK(
		ps.Name("when-new-block-saved-in-consensus-state-func"), cmd.pWhenNewBlockSavedInConsensusStateFunc)
	_ = pps.POK(launch.PNameStates).
		PreAddOK(PNameOperationProcessorsMap, POperationProcessorsMap)
	_ = pps.POK(PNameDigest).
		PostAddOK(PNameDigestAPIHandlers, cmd.pDigestAPIHandlers)
	_ = pps.POK(PNameDigester).
		PostAddOK(PNameDigesterFollowUp, PdigesterFollowUp)

	_ = pps.SetLogging(log)

	log.Log().Debug().Interface("process", pps.Verbose()).Msg("process ready")

	pctx, err := pps.Run(pctx) //revive:disable-line:modifies-parameter
	defer func() {
		log.Log().Debug().Interface("process", pps.Verbose()).Msg("process will be closed")

		if _, err = pps.Close(pctx); err != nil {
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

	return cmd.run(pctx)
}

var errHoldStop = util.NewError("hold stop")

func (cmd *RunCommand) run(pctx context.Context) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	exitch := make(chan error)

	cmd.exitf = func(err error) {
		exitch <- err
	}

	stopstates := func() {}

	if !cmd.holded {
		deferred, err := cmd.runStates(ctx, pctx)
		if err != nil {
			return err
		}

		stopstates = deferred
	}

	select {
	case <-ctx.Done(): // NOTE graceful stop
		return ctx.Err()
	case err := <-exitch:
		if errors.Is(err, errHoldStop) {
			stopstates()

			<-ctx.Done()

			return ctx.Err()
		}

		return err
	}
}

func (cmd *RunCommand) runStates(ctx, pctx context.Context) (func(), error) {
	var discoveries *util.Locked
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

	go func() {
		cmd.exitf(<-states.Wait(ctx))
	}()

	return func() {
		if err := states.Hold(); err != nil && !errors.Is(err, util.ErrDaemonAlreadyStopped) {
			cmd.log.Error().Err(err).Msg("failed to stop states")

			return
		}

		cmd.log.Debug().Msg("states stopped")
	}, nil
}

func (cmd *RunCommand) pWhenNewBlockSavedInConsensusStateFunc(pctx context.Context) (context.Context, error) {
	var log *logging.Logging
	var db isaac.Database
	var params *isaac.LocalParams
	var ballotbox *isaacstates.Ballotbox
	var nodeinfo *isaacnetwork.NodeInfoUpdater
	var di *digest.Digester

	if err := util.LoadFromContextOK(pctx,
		launch.LoggingContextKey, &log,
		launch.CenterDatabaseContextKey, &db,
		launch.LocalParamsContextKey, &params,
		launch.BallotboxContextKey, &ballotbox,
		launch.NodeInfoContextKey, &nodeinfo,
		ContextValueDigester, &di,
	); err != nil {
		return pctx, err
	}
	g := cmd.whenBlockSaved(db, di)

	f := func(height base.Height) {
		launch.WhenNewBlockSavedInConsensusStateFunc(params, ballotbox, db, nodeinfo)(height)
		g(pctx)

		l := log.Log().With().Interface("height", height).Logger()
		l.Debug().Msg("new block saved")

		if cmd.Hold.IsSet() && height == cmd.Hold.Height() {
			l.Debug().Msg("will be stopped by hold")

			cmd.exitf(errHoldStop.Call())

			return
		}
	}

	//revive:disable-next-line:modifies-parameter
	pctx = context.WithValue(pctx, launch.WhenNewBlockSavedInConsensusStateFuncContextKey, f)

	return pctx, nil
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
			return ctx, errors.Errorf("last BlockMap not found")
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

func (cmd *RunCommand) runHTTPState(bind string) error {
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return errors.Wrap(err, "failed to parse --http-state")
	}

	mux := http.NewServeMux()
	if err := statsviz.Register(mux); err != nil {
		return errors.Wrap(err, "failed to register statsviz for http-state")
	}

	cmd.log.Debug().Stringer("bind", addr).Msg("statsviz started")

	go func() {
		_ = http.ListenAndServe(addr.String(), mux)
	}()

	return nil
}

func (cmd *RunCommand) pDigestAPIHandlers(ctx context.Context) (context.Context, error) {
	var params base.LocalParams
	var local base.LocalNode

	util.LoadFromContextOK(ctx,
		launch.LocalContextKey, &local,
		launch.LocalParamsContextKey, &params,
	)

	var design DigestDesign
	if err := util.LoadFromContext(ctx, ContextValueDigestDesign, &design); err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return ctx, nil
		}

		return nil, err
	}

	cache, err := cmd.loadCache(ctx, design)
	if err != nil {
		return ctx, err
	}

	handlers, err := cmd.setDigestHandlers(ctx, local, params, design, cache)
	if err != nil {
		return ctx, err
	}

	if err := handlers.Initialize(); err != nil {
		return ctx, err
	}

	var dnt *digest.HTTP2Server
	if err := util.LoadFromContext(ctx, ContextValueDigestNetwork, &dnt); err != nil {
		return ctx, err
	}
	dnt.SetRouter(handlers.Router())

	return ctx, nil
}

func (cmd *RunCommand) loadCache(_ context.Context, design DigestDesign) (digest.Cache, error) {
	c, err := digest.NewCacheFromURI(design.Cache().String())
	if err != nil {
		cmd.log.Err(err).Str("cache", design.Cache().String()).Msg("failed to connect cache server")
		cmd.log.Warn().Msg("instead of remote cache server, internal mem cache can be available, `memory://`")

		return nil, err
	}
	return c, nil
}

func (cmd *RunCommand) setDigestHandlers(
	ctx context.Context,
	local base.LocalNode,
	params base.LocalParams,
	design DigestDesign,
	cache digest.Cache,
) (*digest.Handlers, error) {
	/*
			var nt network.Server
			if err := process.LoadNetworkContextValue(ctx, &nt); err != nil {
				return nil, err
			}

		var cp *currency.CurrencyPool
		if err := LoadCurrencyPoolContextValue(ctx, &cp); err != nil {
			return nil, err
		}

			handlers := digest.NewHandlers(ctx, conf.NetworkID(), encs, jenc, st, cache, cp).
				SetNodeInfoHandler(nt.NodeInfoHandler())
	*/
	var st *digest.Database
	if err := util.LoadFromContext(ctx, ContextValueDigestDatabase, &st); err != nil {
		return nil, err
	}

	handlers := digest.NewHandlers(ctx, params.NetworkID(), encs, enc, st, cache)
	i, err := cmd.setDigestSendHandler(ctx, local, params, handlers)
	if err != nil {
		return nil, err
	}
	handlers = i

	return handlers, nil
}

func (cmd *RunCommand) setDigestSendHandler(
	ctx context.Context,
	local base.LocalNode,
	params base.LocalParams,
	handlers *digest.Handlers,
) (*digest.Handlers, error) {
	var memberlist *quicmemberlist.Memberlist
	if err := util.LoadFromContextOK(ctx, launch.MemberlistContextKey, &memberlist); err != nil {
		return nil, err
	}

	client := launch.NewNetworkClient( //nolint:gomnd //...
		encs, enc, time.Second*2,
		base.NetworkID([]byte(params.NetworkID())),
	)

	handlers = handlers.SetSend(
		NewSendHandler(local.Privatekey(), params.NetworkID(), func() (*isaacnetwork.QuicstreamClient, *quicmemberlist.Memberlist, error) { // nolint:contextcheck
			return client, memberlist, nil
		}),
	)

	// var memberlist *quicmemberlist.Memberlist
	// if err := util.LoadFromContextOK(ctx, launch.MemberlistContextKey, &memberlist); err != nil {
	// 	return nil, err
	// }

	// handlers = handlers.SetSend(
	// 	NewSendHandler(local.Privatekey(), params.NetworkID(), func() (*quicmemberlist.Memberlist, error) { // nolint:contextcheck
	// 		return memberlist, nil
	// 	}),
	// )

	cmd.log.Debug().Msg("send handler attached")

	return handlers, nil
}
