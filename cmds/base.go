package cmds

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/localtime"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/ProtoconNet/mitum2/util/ps"
	"github.com/rs/zerolog"
)

type BaseCommand struct {
	Encoder  encoder.Encoder   `kong:"-"`
	Encoders *encoder.Encoders `kong:"-"`
	Log      *zerolog.Logger   `kong:"-"`
	Out      io.Writer         `kong:"-"`
}

func (cmd *BaseCommand) prepare(pctx context.Context) (context.Context, error) {
	cmd.Out = os.Stdout
	pps := ps.NewPS("cmd")

	_ = pps.
		AddOK(launch.PNameEncoder, PEncoder, nil)

	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)

	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return pctx, err
	}

	cmd.Log = log.Log()

	pctx, err := pps.Run(pctx) //revive:disable-line:modifies-parameter
	if err != nil {
		return pctx, err
	}

	if err := util.LoadFromContextOK(pctx, launch.EncodersContextKey, &cmd.Encoders); err != nil {
		return pctx, err
	}

	cmd.Encoder = cmd.Encoders.JSON()

	return pctx, nil
}

func (cmd *BaseCommand) print(f string, a ...interface{}) {
	_, _ = fmt.Fprintf(cmd.Out, f, a...)
	_, _ = fmt.Fprintln(cmd.Out)
}

func PAddHinters(pctx context.Context) (context.Context, error) {
	e := util.StringError("add hinters")

	var encs *encoder.Encoders
	if err := util.LoadFromContextOK(pctx, launch.EncodersContextKey, &encs); err != nil {
		return pctx, e.Wrap(err)
	}

	if err := LoadHinters(encs); err != nil {
		return pctx, e.Wrap(err)
	}

	return pctx, nil
}

type OperationFlags struct {
	Privatekey PrivatekeyFlag `arg:"" name:"privatekey" help:"privatekey to sign operation" required:"true"`
	Token      string         `help:"token for operation" optional:""`
	NetworkID  NetworkIDFlag  `name:"network-id" help:"network-id" required:"true" default:"${network_id}"`
	Pretty     bool           `name:"pretty" help:"pretty format"`
}

func (op *OperationFlags) IsValid([]byte) error {
	if len(op.Token) < 1 {
		op.Token = localtime.Now().UTC().String()
	}

	return op.NetworkID.NetworkID().IsValid(nil)
}
