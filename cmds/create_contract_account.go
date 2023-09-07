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
	Sender      AddressFlag          `arg:"" name:"sender" help:"sender address" required:"true"`
	Threshold   uint                 `help:"threshold for keys (default: ${create_contract_account_threshold})" default:"${create_contract_account_threshold}"` // nolint
	Keys        []KeyFlag            `name:"key" help:"key for new account (ex: \"<public key>,<weight>\")" sep:"@"`
	Amounts     []CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	AddressType string               `help:"address type for new account select mitum or ether" default:"mitum"`
	sender      base.Address
	keys        types.AccountKeys
}

func (cmd *CreateContractAccountCommand) Run(pctx context.Context) error { // nolint:dupl
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

func (cmd *CreateContractAccountCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Sender.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	}
	cmd.sender = a

	if len(cmd.Keys) < 1 {
		return errors.Errorf("--key must be given at least one")
	}

	if len(cmd.Amounts) < 1 {
		return errors.Errorf("empty currency-amount, must be given at least one")
	}

	{
		ks := make([]types.AccountKey, len(cmd.Keys))
		for i := range cmd.Keys {
			ks[i] = cmd.Keys[i].Key
		}

		var kys types.AccountKeys
		switch {
		case cmd.AddressType == "ether":
			if kys, err = types.NewEthAccountKeys(ks, cmd.Threshold); err != nil {
				return err
			}
		default:
			if kys, err = types.NewBaseAccountKeys(ks, cmd.Threshold); err != nil {
				return err
			}
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

	ams := make([]types.Amount, len(cmd.Amounts))
	for i := range cmd.Amounts {
		a := cmd.Amounts[i]
		am := types.NewAmount(a.Big, a.CID)
		if err := am.IsValid(nil); err != nil {
			return nil, err
		}

		ams[i] = am
	}

	addrType := types.AddressHint.Type()

	if cmd.AddressType == "ether" {
		addrType = types.EthAddressHint.Type()
	}

	item := extension.NewCreateContractAccountItemMultiAmounts(cmd.keys, ams, addrType)
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
