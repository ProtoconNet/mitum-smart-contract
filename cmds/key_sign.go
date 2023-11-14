package cmds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/launch"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type KeySignCommand struct {
	BaseCommand
	KeyString string             `arg:"" name:"privatekey" help:"privatekey string"`
	NetworkID string             `arg:"" name:"network-id" help:"network-id"`
	Body      *os.File           `arg:"" help:"body"`
	Node      launch.AddressFlag `help:"node address"`
	Token     string             `help:"set fact token"`
	priv      base.Privatekey
	networkID base.NetworkID
}

func (cmd *KeySignCommand) Run(pctx context.Context) error {
	if err := cmd.prepare(pctx); err != nil {
		return err
	}

	cmd.Log.Debug().
		Str("privatekey", cmd.KeyString).
		Str("network_id", cmd.NetworkID).
		Stringer("node", cmd.Node.Address()).
		Msg("flags")

	defer func() {
		_ = cmd.Body.Close()
	}()

	var ptr interface{}

	switch j, err := cmd.loadBody(); {
	case err != nil:
		return err
	default:
		ptr = j
	}

	if _, ok := ptr.(base.NodeSigner); ok && cmd.Node.Address() == nil {
		return errors.Errorf("--node is missing")
	}

	if err := cmd.updateToken(ptr); err != nil {
		return err
	}

	if err := cmd.sign(ptr); err != nil {
		return err
	}

	cmd.Log.Debug().Msg("successfully sign")

	b, err := util.MarshalJSONIndent(ptr)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(os.Stdout, string(b))

	return nil
}

func (cmd *KeySignCommand) prepare(pctx context.Context) error {
	if _, err := cmd.BaseCommand.prepare(pctx); err != nil {
		return err
	}

	switch key, err := base.DecodePrivatekeyFromString(cmd.KeyString, cmd.Encoder); {
	case err != nil:
		return err
	default:
		if err := key.IsValid(nil); err != nil {
			return err
		}

		cmd.priv = key
	}

	cmd.networkID = base.NetworkID([]byte(cmd.NetworkID))

	return cmd.networkID.IsValid(nil)
}

func (cmd *KeySignCommand) loadBody() (interface{}, error) {
	var body []byte

	switch i, err := io.ReadAll(cmd.Body); {
	case err != nil:
		return nil, errors.WithStack(err)
	default:
		body = i
	}

	var u map[string]interface{}
	if err := util.UnmarshalJSON(body, &u); err != nil {
		return nil, err
	}

	switch i, err := util.MarshalJSONIndent(u); {
	case err != nil:
		return nil, err
	default:
		_, _ = fmt.Fprintln(os.Stderr, string(i))
	}

	cmd.Log.Debug().Str("raw_body", string(body)).Msg("read body")

	elem, err := cmd.Encoder.Decode(body)
	if err != nil {
		return nil, err
	}

	if elem == nil {
		return nil, errors.Errorf("load body")
	}

	ptr := reflect.New(reflect.ValueOf(elem).Type()).Interface()

	if err := util.ReflectSetInterfaceValue(elem, ptr); err != nil {
		return nil, err
	}

	cmd.Log.Debug().Str("body_type", fmt.Sprintf("%T", elem)).Msg("body loaded")

	return ptr, nil
}

func (cmd *KeySignCommand) updateToken(ptr interface{}) error {
	var token base.Token

	if i, ok := ptr.(base.Facter); ok {
		if j, ok := i.Fact().(base.Tokener); ok {
			token = j.Token()
		}
	}

	cmd.Log.Debug().Interface("body_token", token).Interface("new_token", []byte(cmd.Token)).Msg("tokens")

	switch {
	case len(token) < 1:
		if len(cmd.Token) < 1 {
			return errors.Errorf("empty token")
		}

		token = base.Token([]byte(cmd.Token))
	case len(cmd.Token) > 0:
		if !bytes.Equal([]byte(cmd.Token), token) {
			return errors.Errorf("different token found")
		}

		cmd.Log.Debug().Msg("same token given")
	}

	if i, ok := ptr.(base.TokenSetter); ok {
		if err := i.SetToken(token); err != nil {
			return err
		}

		cmd.Log.Debug().Interface("new_token", token).Msg("token updated")
	}

	return nil
}

func (cmd *KeySignCommand) sign(ptr interface{}) error {
	var sign func() error

	switch t := ptr.(type) {
	case base.NodeSigner:
		sign = func() error {
			return t.NodeSign(cmd.priv, cmd.networkID, cmd.Node.Address())
		}
	case base.Signer:
		sign = func() error {
			return t.Sign(cmd.priv, cmd.networkID)
		}
	default:
		return errors.Errorf("it's not Signer, %T", ptr)
	}

	if err := sign(); err != nil {
		return err
	}

	if i, ok := ptr.(util.IsValider); ok {
		if err := i.IsValid(cmd.networkID); err != nil {
			return err
		}
	}

	return nil
}
