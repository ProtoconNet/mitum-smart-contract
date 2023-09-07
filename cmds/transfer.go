package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"

	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type TransferCommand struct {
	BaseCommand
	OperationFlags
	Sender   AddressFlag          `arg:"" name:"sender" help:"sender address" required:"true"`
	Receiver AddressFlag          `arg:"" name:"receiver" help:"receiver address" required:"true"`
	Amounts  []CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	sender   base.Address
	receiver base.Address
}

func (cmd *TransferCommand) Run(pctx context.Context) error {
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

func (cmd *TransferCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Amounts) < 1 {
		return errors.Errorf("empty currency-amount, must be given at least one")
	}

	if sender, err := cmd.Sender.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	} else if receiver, err := cmd.Receiver.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	} else {
		cmd.sender = sender
		cmd.receiver = receiver
	}

	return nil
}

func (cmd *TransferCommand) createOperation() (base.Operation, error) { // nolint:dupl
	var items []currency.TransferItem

	ams := make([]types.Amount, len(cmd.Amounts))
	for i := range cmd.Amounts {
		a := cmd.Amounts[i]
		am := types.NewAmount(a.Big, a.CID)
		if err := am.IsValid(nil); err != nil {
			return nil, err
		}

		ams[i] = am
	}

	item := currency.NewTransferItemMultiAmounts(cmd.receiver, ams)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := currency.NewTransferFact([]byte(cmd.Token), cmd.sender, items)

	op, err := currency.NewTransfer(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create transfer operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "create transfer operation")
	}

	return op, nil
}
