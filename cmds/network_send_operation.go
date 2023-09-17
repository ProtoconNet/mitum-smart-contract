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
	Client   *isaacnetwork.BaseClient `kong:"-"`
	ClientID string                   `name:"client-id" help:"client id"`
}

func (cmd *BaseNetworkClientCommand) Prepare(pctx context.Context) error {
	if _, err := cmd.BaseCommand.prepare(pctx); err != nil {
		return err
	}

	if len(cmd.NetworkID) < 1 {
		return errors.Errorf(`expected "<network-id>"`)
	}

	if cmd.Timeout < 1 {
		cmd.Timeout = isaac.DefaultTimeoutRequest * 2
	}

	connectionPool, err := launch.NewConnectionPool(
		1<<9, //nolint:gomnd //...
		base.NetworkID(cmd.NetworkID),
		nil,
	)
	if err != nil {
		return err
	}

	cmd.Client = isaacnetwork.NewBaseClient(
		cmd.Encoders, cmd.Encoder,
		connectionPool.Dial,
		connectionPool.CloseAll,
	)

	cmd.Client.SetClientID(cmd.ClientID)

	cmd.Log.Debug().
		Stringer("remote", cmd.Remote).
		Stringer("timeout", cmd.Timeout).
		Str("network_id", cmd.NetworkID).
		Str("client_id", cmd.ClientID).
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
	//revive:disable:nested-structs
	NodeInfo      launchcmd.NetworkClientNodeInfoCommand     `cmd:"" name:"node-info" help:"remote node info"`
	SendOperation NetworkClientSendOperationCommand          `cmd:"" name:"send-operation" help:"send operation"`
	State         launchcmd.NetworkClientStateCommand        `cmd:"" name:"state" help:"get state"`
	LastBlockMap  launchcmd.NetworkClientLastBlockMapCommand `cmd:"" name:"last-blockmap" help:"get last blockmap"`
	Design        struct {
		Read  launchcmd.NetworkClientReadNodeCommand  `cmd:"" name:"read" help:"read design value"`
		Write launchcmd.NetworkClientWriteNodeCommand `cmd:"" name:"write" help:"write design value"`
	} `cmd:"" name:"design" help:""`
	//revive:enable:nested-structs
	//revive:enable:line-length-limit
}

type NetworkClientSendOperationCommand struct { //nolint:govet //...
	BaseNetworkClientCommand
}

func (cmd *NetworkClientSendOperationCommand) Run(pctx context.Context) error {
	if err := cmd.Prepare(pctx); err != nil {
		return err
	}

	defer func() {
		_ = cmd.Client.Close()
	}()

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, cmd.Body); err != nil {
		return errors.WithStack(err)
	}

	var op base.Operation
	if err := encoder.Decode(cmd.Encoder, buf.Bytes(), &op); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(pctx, cmd.Timeout)
	defer cancel()

	switch sent, err := cmd.Client.SendOperation(ctx, cmd.Remote.ConnInfo(), op); {
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
