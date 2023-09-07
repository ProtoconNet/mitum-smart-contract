package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/logging"
)

type INITCommand struct {
	GenesisDesign string `arg:"" name:"genesis design" help:"genesis design" type:"filepath"`
	Vault         string `name:"vault" help:"privatekey path of vault"`
	launch.DesignFlag
	launch.DevFlags `embed:"" prefix:"dev."`
}

func NewINITCommand() INITCommand {
	return INITCommand{}
}

func (cmd *INITCommand) Run(pctx context.Context) error {
	var log *logging.Logging
	if err := util.LoadFromContextOK(pctx, launch.LoggingContextKey, &log); err != nil {
		return err
	}

	nctx := util.ContextWithValues(pctx, map[util.ContextKey]interface{}{
		launch.DesignFlagContextKey:        cmd.DesignFlag,
		launch.DevFlagsContextKey:          cmd.DevFlags,
		launch.GenesisDesignFileContextKey: cmd.GenesisDesign,
		launch.VaultContextKey:             cmd.Vault,
	})

	pps := DefaultINITPS()
	_ = pps.SetLogging(log)

	log.Log().Debug().Interface("process", pps.Verbose()).Msg("process ready")

	_, err := pps.Run(nctx) //revive:disable-line:modifies-parameter
	defer func() {
		log.Log().Debug().Interface("process", pps.Verbose()).Msg("process will be closed")

		if _, err = pps.Close(pctx); err != nil {
			log.Log().Error().Err(err).Msg("close")
		}
	}()

	return err
}
