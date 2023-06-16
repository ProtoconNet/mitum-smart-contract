package cmds

import (
	"bytes"
	"context"
	"io"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

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
