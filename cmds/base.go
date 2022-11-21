package cmds

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/launch"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/localtime"
	"github.com/spikeekips/mitum/util/logging"
	"github.com/spikeekips/mitum/util/ps"
)

type baseCommand struct {
	enc  *jsonenc.Encoder
	encs *encoder.Encoders
	log  *zerolog.Logger
	Out  io.Writer `kong:"-"`
}

func NewbaseCommand() *baseCommand {
	return &baseCommand{
		Out: os.Stdout,
	}
}

func (cmd *baseCommand) prepare(pctx context.Context) (context.Context, error) {
	pps := ps.NewPS("cmd")

	_ = pps.
		AddOK(launch.PNameEncoder, PEncoder, nil)

	_ = pps.POK(launch.PNameEncoder).
		PostAddOK(launch.PNameAddHinters, PAddHinters)

	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return pctx, err
	}

	cmd.log = log.Log()

	pctx, err := pps.Run(pctx) //revive:disable-line:modifies-parameter
	if err != nil {
		return pctx, err
	}

	return pctx, util.LoadFromContextOK(pctx,
		launch.EncodersContextKey, &cmd.encs,
		launch.EncoderContextKey, &cmd.enc,
	)
}

func (cmd *baseCommand) print(f string, a ...interface{}) {
	_, _ = fmt.Fprintf(cmd.Out, f, a...)
	_, _ = fmt.Fprintln(cmd.Out)
}

func PAddHinters(ctx context.Context) (context.Context, error) {
	e := util.StringErrorFunc("failed to add hinters")

	var enc encoder.Encoder
	if err := util.LoadFromContextOK(ctx, launch.EncoderContextKey, &enc); err != nil {
		return ctx, e(err, "")
	}

	if err := LoadHinters(enc); err != nil {
		return ctx, e(err, "")
	}

	return ctx, nil
}

type OperationFlags struct {
	Privatekey PrivatekeyFlag `arg:"" name:"privatekey" help:"privatekey to sign operation" required:"true"`
	Token      string         `help:"token for operation" optional:""`
	NetworkID  NetworkIDFlag  `name:"network-id" help:"network-id" required:"true"`
	Memo       string         `name:"memo" help:"memo"`
	Pretty     bool           `name:"pretty" help:"pretty format"`
}

func (op *OperationFlags) IsValid([]byte) error {
	if len(op.Token) < 1 {
		op.Token = localtime.UTCNow().String()
	}

	return op.NetworkID.NetworkID().IsValid(nil)
}
