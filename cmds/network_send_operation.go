package cmds

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ProtoconNet/mitum2/isaac"
	isaacnetwork "github.com/ProtoconNet/mitum2/isaac/network"
	"github.com/ProtoconNet/mitum2/launch"
	launchcmd "github.com/ProtoconNet/mitum2/launch/cmd"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"io"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

type BaseNetworkClientCommand struct { //nolint:govet //...
	BaseCommand
	launchcmd.BaseNetworkClientNodeInfoFlags
	Client *isaacnetwork.QuicstreamClient `kong:"-"`
}

func (cmd *BaseNetworkClientCommand) Prepare(pctx context.Context) error {
	if _, err := cmd.BaseCommand.prepare(pctx); err != nil {
		return err
	}

	if len(cmd.NetworkID) < 1 {
		return errors.Errorf(`expected "<network-id>"`)
	}

	if _, err := cmd.Remote.ConnInfo(); err != nil {
		return err
	}

	if cmd.Timeout < 1 {
		cmd.Timeout = isaac.DefaultTimeoutRequest * 2
	}

	cmd.Client = launch.NewNetworkClient(cmd.Encoders, cmd.Encoder, base.NetworkID(cmd.NetworkID))

	cmd.Log.Debug().
		Stringer("remote", cmd.Remote).
		Stringer("timeout", cmd.Timeout).
		Str("network_id", cmd.NetworkID).
		Bool("has_body", cmd.Body != nil).
		Msg("flags")

	return nil
}

func (cmd *BaseNetworkClientCommand) Print(v interface{}, out io.Writer) error {
	l := cmd.Log.Debug().
		Str("type", fmt.Sprintf("%T", v))

	if ht, ok := v.(hint.Hinter); ok {
		l = l.Stringer("hint", ht.Hint())
	}

	l.Msg("body loaded")

	b, err := util.MarshalJSONIndent(v)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, string(b))

	return errors.WithStack(err)
}

type NetworkClientCommand struct { //nolint:govet //...
	//revive:disable:line-length-limit
	NodeInfo          launchcmd.NetworkClientNodeInfoCommand          `cmd:"" name:"node-info" help:"remote node info"`
	SendOperation     NetworkClientSendOperationCommand               `cmd:"" name:"send-operation" help:"send operation"`
	State             launchcmd.NetworkClientStateCommand             `cmd:"" name:"state" help:"get state"`
	LastBlockMap      launchcmd.NetworkClientLastBlockMapCommand      `cmd:"" name:"last-blockmap" help:"get last blockmap"`
	SetAllowConsensus launchcmd.NetworkClientSetAllowConsensusCommand `cmd:"" name:"set-allow-consensus" help:"set to enter consensus"`
	//revive:enable:line-length-limit
}

type NetworkClientSendOperationCommand struct { //nolint:govet //...
	BaseNetworkClientCommand
}

func (cmd *NetworkClientSendOperationCommand) Run(pctx context.Context) error {
	if err := cmd.Prepare(pctx); err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, cmd.Body); err != nil {
		return errors.WithStack(err)
	}

	var op base.Operation
	if err := encoder.Decode(cmd.Encoder, buf.Bytes(), &op); err != nil {
		return err
	}

	ci, _ := cmd.Remote.ConnInfo()

	ctx, cancel := context.WithTimeout(pctx, cmd.Timeout)
	defer cancel()

	switch sent, err := cmd.Client.SendOperation(ctx, ci, op); {
	case err != nil:
		cmd.Log.Error().Err(err).Msg("not sent")

		return err
	case !sent:
		cmd.Log.Error().Msg("not sent")
	default:
		cmd.Log.Info().Msg("sent")
	}

	return nil
}
