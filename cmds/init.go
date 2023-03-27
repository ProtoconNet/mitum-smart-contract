package cmds

import (
	"context"
	"os"

	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util/encoder"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

var (
	encs *encoder.Encoders
	enc  *jsonenc.Encoder
)

func init() {
	pctx := context.Background()
	baseFlags := launch.BaseFlags{
		LoggingFlags: launch.LoggingFlags{
			Out:    launch.LogOutFlag("stdout"),
			Format: "terminal",
		},
	}
	pctx = context.WithValue(pctx, launch.FlagsContextKey, baseFlags)
	log, err := launch.SetupLoggingFromFlags(baseFlags.LoggingFlags)
	if err != nil {
		panic(err)
	}

	pctx = context.WithValue(pctx, launch.LoggingContextKey, log) //revive:disable-line:modifies-parameter

	cmd := baseCommand{
		Out: os.Stdout,
	}

	if _, err := cmd.prepare(pctx); err != nil {
		panic(err)
	} else {
		encs = cmd.encs
		enc = cmd.enc
	}
}
