package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type UpdateKeyCommand struct {
	BaseCommand
	OperationFlags
	Target    AddressFlag    `arg:"" name:"target" help:"target address" required:"true"`
	Threshold uint           `help:"threshold for keys (default: ${create_account_threshold})" default:"${create_account_threshold}"` // nolint
	Keys      []KeyFlag      `name:"key" help:"key for new account (ex: \"<public key>,<weight>\")" sep:"@"`
	Currency  CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	target    base.Address
	keys      types.BaseAccountKeys
}

func (cmd *UpdateKeyCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	encs = cmd.Encoders
	enc = cmd.Encoder

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	op, err := cmd.createOperation()
	if err != nil {
		return err
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *UpdateKeyCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Target.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Target.String())
	}
	cmd.target = a

	if len(cmd.Keys) < 1 {
		return errors.Errorf("--key must be given at least one")
	}

	{
		ks := make([]types.AccountKey, len(cmd.Keys))
		for i := range cmd.Keys {
			ks[i] = cmd.Keys[i].Key
		}

		if kys, err := types.NewBaseAccountKeys(ks, cmd.Threshold); err != nil {
			return err
		} else if err := kys.IsValid(nil); err != nil {
			return err
		} else {
			cmd.keys = kys
		}
	}

	return nil
}

func (cmd *UpdateKeyCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	fact := currency.NewUpdateKeyFact([]byte(cmd.Token), cmd.target, cmd.keys, cmd.Currency.CID)

	op, err := currency.NewUpdateKey(fact)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create update-key operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create update-key operation")
	}

	return op, nil
}
