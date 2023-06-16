package cmds

import (
	"context"
	"fmt"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"os"
	"strings"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type KeyNewCommand struct {
	BaseCommand
	Seed    string `arg:"" name:"seed" optional:"" help:"seed for generating key"`
	KeyType string `help:"select btc or ether" default:"btc"`
}

func NewKeyNewCommand() KeyNewCommand {
	cmd := NewBaseCommand()
	return KeyNewCommand{
		BaseCommand: *cmd,
	}
}

func (cmd *KeyNewCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	cmd.Log.Debug().
		Str("seed", cmd.Seed).
		Msg("flags")

	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	var key base.Privatekey

	switch {
	case len(cmd.Seed) > 0:
		if len(strings.TrimSpace(cmd.Seed)) < 1 {
			cmd.Log.Warn().Msg("seed consists with empty spaces")
		}
		if len(cmd.KeyType) > 0 && cmd.KeyType == "ether" {
			i, err := types.NewMEPrivatekeyFromSeed(cmd.Seed)
			if err != nil {
				return err
			}
			key = i
		} else {
			i, err := base.NewMPrivatekeyFromSeed(cmd.Seed)
			if err != nil {
				return err
			}
			key = i
		}

	default:
		if len(cmd.KeyType) > 0 && cmd.KeyType == "ether" {
			key = types.NewMEPrivatekey()
		} else {
			key = base.NewMPrivatekey()
		}
	}

	o := struct {
		PrivateKey base.PKKey  `json:"privatekey"` //nolint:tagliatelle //...
		Publickey  base.PKKey  `json:"publickey"`
		Hint       interface{} `json:"hint,omitempty"`
		Seed       string      `json:"seed"`
		Type       string      `json:"type"`
	}{
		Seed:       cmd.Seed,
		PrivateKey: key,
		Publickey:  key.Publickey(),
		Type:       "privatekey",
	}

	if hinter, ok := (interface{})(key).(hint.Hinter); ok {
		o.Hint = hinter.Hint()
	}

	b, err := util.MarshalJSONIndent(o)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, string(b))

	return nil
}
