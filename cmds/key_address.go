package cmds

import (
	"context"
	base3 "github.com/ProtoconNet/mitum-currency/v2/base"
	"github.com/alecthomas/kong"

	"github.com/ProtoconNet/mitum2/base"
)

var KeyAddressVars = kong.Vars{
	"create_account_threshold": "100",
}

type KeyAddressCommand struct {
	baseCommand
	Threshold   uint      `arg:"" name:"threshold" help:"threshold for keys (default: ${create_account_threshold})" default:"${create_account_threshold}"` // nolint
	Keys        []KeyFlag `arg:"" name:"key" help:"key for address (ex: \"<public key>,<weight>\")" sep:"@" optional:""`
	AddressType string    `help:"key type for address. select btc or ether" default:"btc"`
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

	ks := make([]base3.AccountKey, len(cmd.Keys))
	for i := range cmd.Keys {
		ks[i] = cmd.Keys[i].Key
	}

	keys, err := base3.NewBaseAccountKeys(ks, cmd.Threshold)
	if err != nil {
		return err
	}

	cmd.log.Debug().Int("number_of_keys", len(ks)).Interface("keys", keys).Msg("keys loaded")

	var a base.Address
	if len(cmd.AddressType) > 0 && cmd.AddressType == "ether" {
		a, err = base3.NewEthAddressFromKeys(keys)
	} else {
		a, err = base3.NewAddressFromKeys(keys)
	}

	if err != nil {
		return err
	}
	cmd.print(a.String())

	return nil
}
