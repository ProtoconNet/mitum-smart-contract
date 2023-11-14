package cmds

import (
	"context"
	"os"

	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

var (
	encs *encoder.Encoders
	enc  encoder.Encoder
)

func init() {
	pctx := context.Background()
	baseFlags := launch.BaseFlags{
		LoggingFlags: launch.LoggingFlags{
			Out:    []launch.LogOutFlag{launch.LogOutFlag("stdout")},
			Format: "terminal",
		},
	}
	pctx = context.WithValue(pctx, launch.FlagsContextKey, baseFlags)
	log, logout, err := launch.SetupLoggingFromFlags(baseFlags.LoggingFlags)
	if err != nil {
		panic(err)
	}

	pctx = context.WithValue(pctx, launch.LoggingContextKey, log)   //revive:disable-line:modifies-parameter
	pctx = context.WithValue(pctx, launch.LogOutContextKey, logout) //revive:disable-line:modifies-parameter

	cmd := BaseCommand{
		Out: os.Stdout,
	}

	if _, err := cmd.prepare(pctx); err != nil {
		panic(err)
	} else {
		encs = cmd.Encoders
		enc = cmd.Encoder
	}
}
