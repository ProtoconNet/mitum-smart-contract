package cmds

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
)

type KeyUpdaterCommand struct {
	baseCommand
	OperationFlags
	Target    AddressFlag    `arg:"" name:"target" help:"target address" required:"true"`
	Threshold uint           `help:"threshold for keys (default: ${create_account_threshold})" default:"${create_account_threshold}"` // nolint
	Keys      []KeyFlag      `name:"key" help:"key for new account (ex: \"<public key>,<weight>\")" sep:"@"`
	Currency  CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	target    base.Address
	keys      currency.BaseAccountKeys
}

func NewKeyUpdaterCommand() KeyUpdaterCommand {
	cmd := NewbaseCommand()
	return KeyUpdaterCommand{
		baseCommand: *cmd,
	}
}

func (cmd *KeyUpdaterCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	encs = cmd.encs
	enc = cmd.enc

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	/*
		sl, err := LoadSealAndAddOperation(
			cmd.Seal.Bytes(),
			cmd.Privatekey,
			cmd.NetworkID.NetworkID(),
			op,
		)
		if err != nil {
			return err
		}
	*/
	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *KeyUpdaterCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Target.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Target.String())
	}
	cmd.target = a

	if len(cmd.Keys) < 1 {
		return errors.Errorf("--key must be given at least one")
	}

	{
		ks := make([]currency.AccountKey, len(cmd.Keys))
		for i := range cmd.Keys {
			ks[i] = cmd.Keys[i].Key
		}

		if kys, err := currency.NewBaseAccountKeys(ks, cmd.Threshold); err != nil {
			return err
		} else if err := kys.IsValid(nil); err != nil {
			return err
		} else {
			cmd.keys = kys
		}
	}

	return nil
}

func (cmd *KeyUpdaterCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	fact := currency.NewKeyUpdaterFact([]byte(cmd.Token), cmd.target, cmd.keys, cmd.Currency.CID)

	op, err := currency.NewKeyUpdater(fact, cmd.Memo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key-updater operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key-updater operation")
	}

	return op, nil
}
