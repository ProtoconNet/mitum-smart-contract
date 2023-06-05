package cmds

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"github.com/ProtoconNet/mitum-currency/v3/digest"
	"github.com/ProtoconNet/mitum2/base"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/network/quicmemberlist"
	"github.com/ProtoconNet/mitum2/network/quicstream"
	mitumutil "github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/pkg/errors"
)

const (
	ProcessNameDigestAPI      = "digest_api"
	ProcessNameStartDigestAPI = "start_digest_api"
	HookNameSetLocalChannel   = "set_local_channel"
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

	var nt *digest.HTTP2Server
	var certs []tls.Certificate
	if design.Network().Bind().Scheme == "https" {
		certs = design.Network().Certs()
	}

	if sv, err := digest.NewHTTP2Server(
		design.Network().Bind().Host,
		design.Network().ConnInfo().URL().Host,
		certs,
	); err != nil {
		return ctx, err
	} else if err := sv.Initialize(); err != nil {
		return ctx, err
	} else {
		nt = sv
	}

	return context.WithValue(ctx, ContextValueDigestNetwork, nt), nil
}

func NewSendHandler(
	priv base.Privatekey,
	networkID base.NetworkID,
	f func() (*isaacnetwork.QuicstreamClient, *quicmemberlist.Memberlist, error),
) func(interface{}) (base.Operation, error) {
	return func(v interface{}) (base.Operation, error) {
		op, ok := v.(base.Operation)
		if !ok {
			return nil, mitumutil.ErrWrongType.Errorf("expected Operation, not %T", v)
		}

		var header = isaacnetwork.NewSendOperationRequestHeader()

		client, memberlist, err := f()

		switch {
		case err != nil:
			return nil, err

		default:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			var nodelist []quicstream.UDPConnInfo
			memberlist.Members(func(node quicmemberlist.Node) bool {
				nodelist = append(nodelist, node.UDPConnInfo())
				return true
			})
			for i := range nodelist {
				buf := bytes.NewBuffer(nil)
				if err := json.NewEncoder(buf).Encode(op); err != nil {
					return nil, err
				} else if buf == nil {
					return nil, errors.Errorf("buffer from json encoding operation is nil")
				}

				response, _, cancelrequest, err := client.Request(ctx, nodelist[i], header, buf)
				if err != nil {
					return op, err
				}
				if response.Err() != nil {
					return op, response.Err()
				}

				defer func() {
					_ = cancelrequest()
				}()
			}
		}

		return op, nil
	}
}
