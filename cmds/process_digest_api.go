package cmds

import (
	"context"
	"crypto/tls"
	"github.com/ProtoconNet/mitum-currency/v3/digest"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
)

func ProcessStartDigestAPI(ctx context.Context) (context.Context, error) {
	var nt *digest.HTTP2Server
	if err := mitumutil.LoadFromContext(ctx, ContextValueDigestNetwork, &nt); err != nil {
		return ctx, err
	}
	if nt == nil {
		return ctx, nil
	}

	return ctx, nt.Start(ctx)
}

func ProcessDigestAPI(ctx context.Context) (context.Context, error) {
	var design DigestDesign
	if err := mitumutil.LoadFromContext(ctx, ContextValueDigestDesign, &design); err != nil {
		return ctx, err
	}

	var log *logging.Logging
	if err := mitumutil.LoadFromContextOK(ctx, launch.LoggingContextKey, &log); err != nil {
		return ctx, err
	}

	if design.Network() == nil {
		log.Log().Debug().Msg("digest api disabled; empty network")

		return ctx, nil
	}

	var st *digest.Database
	if err := mitumutil.LoadFromContextOK(ctx, ContextValueDigestDatabase, &st); err != nil {
		log.Log().Debug().Err(err).Msg("digest api disabled; empty database")

		return ctx, nil
	} else if st == nil {
		log.Log().Debug().Msg("digest api disabled; empty database")

		return ctx, nil
	}

	log.Log().Info().
		Str("bind", design.Network().Bind().String()).
		Str("publish", design.Network().ConnInfo().String()).
		Msg("trying to start http2 server for digest API")

	var params *launch.LocalParams
	var memberList *quicmemberlist.Memberlist
	if err := mitumutil.LoadFromContextOK(ctx,
		launch.LocalParamsContextKey, &params,
		launch.MemberlistContextKey, &memberList,
	); err != nil {
		return nil, err
	}

	connectionPool, err := launch.NewConnectionPool(
		1<<9,
		params.ISAAC.NetworkID(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	client := isaacnetwork.NewBaseClient( //nolint:gomnd //...
		encs, enc,
		connectionPool.Dial,
		connectionPool.CloseAll,
	)

	var nt *digest.HTTP2Server
	var certs []tls.Certificate
	if design.Network().Bind().Scheme == "https" {
		certs = design.Network().Certs()
	}

	if sv, err := digest.NewHTTP2Server(
		design.Network().Bind().Host,
		design.Network().ConnInfo().URL().Host,
		certs,
		enc,
		params.ISAAC.NetworkID(),
	); err != nil {
		return ctx, err
	} else if err := sv.Initialize(); err != nil {
		return ctx, err
	} else {
		nt = sv
	}

	nt = nt.SetNetworkClientFunc(
		func() (*isaacnetwork.BaseClient, *quicmemberlist.Memberlist, error) { // nolint:contextcheck
			return client, memberList, nil
		},
	)

	return context.WithValue(ctx, ContextValueDigestNetwork, nt), nil
}
