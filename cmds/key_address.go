package cmds

import (
	"context"

	"github.com/alecthomas/kong"

	"github.com/spikeekips/mitum-currency/currency"
)

var KeyAddressVars = kong.Vars{
	"create_account_threshold": "100",
}

type KeyAddressCommand struct {
	baseCommand
	Threshold uint      `arg:"" name:"threshold" help:"threshold for keys (default: ${create_account_threshold})" default:"${create_account_threshold}"` // nolint
	Keys      []KeyFlag `arg:"" name:"key" help:"key for address (ex: \"<public key>,<weight>\")" sep:"@" optional:""`
}

func NewKeyAddressCommand() KeyAddressCommand {
	cmd := NewbaseCommand()
	return KeyAddressCommand{
		baseCommand: *cmd,
	}
}

func (cmd *KeyAddressCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	ks := make([]currency.AccountKey, len(cmd.Keys))
	for i := range cmd.Keys {
		ks[i] = cmd.Keys[i].Key
	}

	keys, err := currency.NewBaseAccountKeys(ks, cmd.Threshold)
	if err != nil {
		return err
	}

	cmd.log.Debug().Int("number_of_keys", len(ks)).Interface("keys", keys).Msg("keys loaded")

	a, err := currency.NewAddressFromKeys(keys)
	if err != nil {
		return err
	}
	cmd.print(a.String())

	return nil
}
