package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/types"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type TransferCommand struct {
	BaseCommand
	OperationFlags
	Sender         AddressFlag               `arg:"" name:"sender" help:"sender address" required:"true"`
	ReceiverAmount AddressCurrencyAmountFlag `arg:"" name:"receiver-currency-amount" help:"receiver amount (ex: \"<address>,<currency>,<amount>\") separator @" required:"true"`
	sender         base.Address
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

	if sender, err := cmd.Sender.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	} else {
		cmd.sender = sender
	}

	return nil
}

func (cmd *TransferCommand) createOperation() (base.Operation, error) { // nolint:dupl
	var items []currency.TransferItem
	for i := range cmd.ReceiverAmount.Address() {
		item := currency.NewTransferItemMultiAmounts(cmd.ReceiverAmount.Address()[i], []types.Amount{cmd.ReceiverAmount.Amount()[i]})
		if err := item.IsValid(nil); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	fact := currency.NewTransferFact([]byte(cmd.Token), cmd.sender, items)

	op, err := currency.NewTransfer(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create transfer operation")
	}

	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "create transfer operation")
	}

	if err := op.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return nil, errors.Wrap(err, "create transfer operation")
	}

	return op, nil
}
