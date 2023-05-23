package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ProtoconNet/mitum-currency/v2/cmds"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
)

var (
	Version   = "v0.0.1"
	BuildTime = "-"
	GitBranch = "master"
	GitCommit = "-"
	version   util.Version
)

//revive:disable:nested-structs
type CLI struct { //nolint:govet //...
	launch.BaseFlags
	Import    cmds.ImportCommand    `cmd:"" help:"import from block data"`
	Init      cmds.INITCommand      `cmd:"" help:"init node"`
	Run       cmds.RunCommand       `cmd:"" help:"run node"`
	Operation cmds.OperationCommand `cmd:"" help:"create operation"`
	Network   cmds.NetworkCommand   `cmd:"" help:"network"`
	Key       cmds.KeyCommand       `cmd:"" help:"key"`
	Version   struct{}              `cmd:"" help:"version"`
}

//revive:enable:nested-structs

var flagDefaults = kong.Vars{
	"log_out":                  "stderr",
	"log_format":               "terminal",
	"log_level":                "debug",
	"log_force_color":          "false",
	"design_uri":               launch.DefaultDesignURI,
	"create_account_threshold": "100",
	"safe_threshold":           base.SafeThreshold.String(),
	"network_id":               "mitum",
}

func main() {
	cli := CLI{
		Import:    cmds.NewImportCommand(),
		Init:      cmds.NewINITCommand(),
		Run:       cmds.NewRunCommand(),
		Operation: cmds.NewOperationCommand(),
		Network:   cmds.NewNetworkCommand(),
		Key:       cmds.NewKeyCommand(),
	}
	kctx := kong.Parse(&cli, flagDefaults)

	if err := checkVersion(); err != nil {
		kctx.FatalIfErrorf(err)
	}

	if kctx.Command() == "version" {
		showVersion()

		return
	}
	pctx := context.Background()
	pctx = context.WithValue(pctx, launch.VersionContextKey, version)
	pctx = context.WithValue(pctx, launch.FlagsContextKey, cli.BaseFlags)
	pctx = context.WithValue(pctx, launch.KongContextContextKey, kctx)

	pss := launch.DefaultMainPS()

	switch i, err := pss.Run(pctx); {
	case err != nil:
		kctx.FatalIfErrorf(err)
	default:
		pctx = i

		kctx = kong.Parse(&cli, kong.BindTo(pctx, (*context.Context)(nil)), flagDefaults)
	}

	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		kctx.FatalIfErrorf(err)
	}

	log.Log().Debug().Interface("flags", os.Args).Msg("flags")
	log.Log().Debug().Interface("main_process", pss.Verbose()).Msg("processed")

	if err := func() error {
		defer log.Log().Debug().Msg("stopped")

		return errors.WithStack(kctx.Run(pctx))
	}(); err != nil {
		log.Log().Error().Err(err).Msg("stopped by error")

		kctx.FatalIfErrorf(err)
	}
}

func checkVersion() error {
	if len(Version) < 1 {
		return errors.Errorf("empty version")
	}

	v, err := util.ParseVersion(Version)
	if err != nil {
		return err
	}

	if err := v.IsValid(nil); err != nil {
		return err
	}

	version = v

	return nil
}

func showVersion() {
	_, _ = fmt.Fprintf(os.Stdout, `version: %s
 branch: %s
 commit: %s
  build: %s
`, version, GitBranch, GitCommit, BuildTime)
}
