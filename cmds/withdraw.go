package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"

	"github.com/pkg/errors"

	mitumbase "github.com/ProtoconNet/mitum2/base"
)

type WithdrawCommand struct {
	baseCommand
	OperationFlags
	Sender  AddressFlag          `arg:"" name:"sender" help:"sender address" required:"true"`
	Target  AddressFlag          `arg:"" name:"target" help:"target contract account address" required:"true"`
	Amounts []CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	sender  mitumbase.Address
	target  mitumbase.Address
}

func NewWithdrawCommand() WithdrawCommand {
	cmd := NewbaseCommand()
	return WithdrawCommand{
		baseCommand: *cmd,
	}
}

func (cmd *WithdrawCommand) Run(pctx context.Context) error {
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

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *WithdrawCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Amounts) < 1 {
		return errors.Errorf("empty currency-amount, must be given at least one")
	}

	if sender, err := cmd.Sender.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	} else if target, err := cmd.Target.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid target format, %q", cmd.Target.String())
	} else {
		cmd.sender = sender
		cmd.target = target
	}

	return nil
}

func (cmd *WithdrawCommand) createOperation() (mitumbase.Operation, error) { // nolint:dupl
	var items []extension.WithdrawsItem

	ams := make([]base.Amount, len(cmd.Amounts))
	for i := range cmd.Amounts {
		a := cmd.Amounts[i]
		am := base.NewAmount(a.Big, a.CID)
		if err := am.IsValid(nil); err != nil {
			return nil, err
		}

		ams[i] = am
	}

	item := extension.NewWithdrawsItemMultiAmounts(cmd.target, ams)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := extension.NewWithdrawsFact([]byte(cmd.Token), cmd.sender, items)

	op, err := extension.NewWithdraws(fact)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create withdraws operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create withdraws operation")
	}

	return op, nil
}
