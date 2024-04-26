package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/alecthomas/kong"
)

var KeyAddressVars = kong.Vars{
	"create_account_threshold": "100",
}

type KeyAddressCommand struct {
	BaseCommand
	Threshold   uint    `arg:"" name:"threshold" help:"threshold for keys (default: ${create_account_threshold})" default:"${create_account_threshold}"` // nolint
	Key         KeyFlag `arg:"" name:"key" help:"key for address (ex: \"<public key>,<weight>\") separator @" optional:""`
	AddressType string  `help:"key type for address. select mitum or ether" default:"mitum"`
}

func (cmd *KeyAddressCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	ks := make([]types.AccountKey, len(cmd.Key.Values))
	for i := range cmd.Key.Values {
		ks[i] = cmd.Key.Values[i]
	}

	var a base.Address
	var keys types.AccountKeys
	var err error
	if cmd.AddressType == "ether" {
		keys, err = types.NewEthAccountKeys(ks, cmd.Threshold)
		if err != nil {
			return err
		}

		cmd.Log.Debug().Int("number_of_keys", len(ks)).Interface("keys", keys).Msg("keys loaded")

		a, err = types.NewEthAddressFromKeys(keys)
		if err != nil {
			return err
		}
	} else {
		keys, err = types.NewBaseAccountKeys(ks, cmd.Threshold)
		if err != nil {
			return err
		}

		cmd.Log.Debug().Int("number_of_keys", len(ks)).Interface("keys", keys).Msg("keys loaded")

		a, err = types.NewAddressFromKeys(keys)
		if err != nil {
			return err
		}
	}

	cmd.print(a.String())

	return nil
}
