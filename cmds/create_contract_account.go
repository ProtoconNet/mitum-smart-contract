package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type CreateContractAccountCommand struct {
	BaseCommand
	OperationFlags
	Sender    AddressFlag        `arg:"" name:"sender" help:"sender address" required:"true"`
	Threshold uint               `help:"threshold for keys (default: ${create_contract_account_threshold})" default:"${create_contract_account_threshold}"` // nolint
	Key       KeyFlag            `name:"key" help:"key for new account (ex: \"<public key>,<weight>\") separator @"`
	Amount    CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	sender    base.Address
	keys      types.AccountKeys
}

func (cmd *CreateContractAccountCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	//encs = cmd.Encoders
	//enc = cmd.Encoder

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

func (cmd *CreateContractAccountCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Sender.Encode(cmd.Encoder)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	}
	cmd.sender = a

	{
		ks := make([]types.AccountKey, len(cmd.Key.Values))
		for i := range cmd.Key.Values {
			ks[i] = cmd.Key.Values[i]
		}

		var kys types.AccountKeys
		if kys, err = types.NewBaseAccountKeys(ks, cmd.Threshold); err != nil {
			return err
		}

		if err := kys.IsValid(nil); err != nil {
			return err
		} else {
			cmd.keys = kys
		}
	}

	return nil
}

func (cmd *CreateContractAccountCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	var items []extension.CreateContractAccountItem

	ams := make([]types.Amount, 1)
	am := types.NewAmount(cmd.Amount.Big, cmd.Amount.CID)
	if err := am.IsValid(nil); err != nil {
		return nil, err
	}

	ams[0] = am

	item := extension.NewCreateContractAccountItemMultiAmounts(cmd.keys, ams)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := extension.NewCreateContractAccountFact([]byte(cmd.Token), cmd.sender, items)

	op, err := extension.NewCreateContractAccount(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create create-contract-account operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "create create-contract-account operation")
	}

	return op, nil
}
