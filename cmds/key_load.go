package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

type KeyLoadCommand struct {
	BaseCommand
	KeyString string `arg:"" name:"key string" help:"key string"`
}

func (cmd *KeyLoadCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	cmd.Log.Debug().
		Str("key_string", cmd.KeyString).
		Msg("flags")

	if len(cmd.KeyString) < 1 {
		return errors.Errorf("empty key string")
	}

	if key, err := base.DecodePrivatekeyFromString(cmd.KeyString, cmd.Encoder); err == nil {
		o := struct {
			PrivateKey base.PKKey  `json:"privatekey"` //nolint:tagliatelle //...
			Publickey  base.PKKey  `json:"publickey"`
			Hint       interface{} `json:"hint,omitempty"`
			String     string      `json:"string"`
			Type       string      `json:"type"`
		}{
			String:     cmd.KeyString,
			PrivateKey: key,
			Publickey:  key.Publickey(),
			Type:       "privatekey",
		}

		if hinter, ok := key.(hint.Hinter); ok {
			o.Hint = hinter.Hint()
		}

		b, err := util.MarshalJSONIndent(o)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, string(b))

		return nil
	}

	if key, err := base.DecodePublickeyFromString(cmd.KeyString, cmd.Encoder); err == nil {
		o := struct {
			Publickey base.PKKey  `json:"publickey"`
			Hint      interface{} `json:"hint,omitempty"`
			String    string      `json:"string"`
			Type      string      `json:"type"`
		}{
			String:    cmd.KeyString,
			Publickey: key,
			Type:      "publickey",
		}

		if hinter, ok := key.(hint.Hinter); ok {
			o.Hint = hinter.Hint()
		}

		b, err := util.MarshalJSONIndent(o)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(os.Stdout, string(b))

		return nil
	}

	return nil
}
