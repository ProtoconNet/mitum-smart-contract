package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type UpdateHandlerCommand struct {
	BaseCommand
	OperationFlags
	Sender   AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract AddressFlag    `arg:"" name:"contract" help:"target contract account address" required:"true"`
	Currency CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	Handlers []AddressFlag  `arg:"" name:"handlers" help:"handlers"`
	sender   base.Address
	target   base.Address
}

func (cmd *UpdateHandlerCommand) Run(pctx context.Context) error {
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

func (cmd *UpdateHandlerCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Handlers) < 1 {
		return errors.Errorf("Empty handlers, must be given at least one")
	}

	if sender, err := cmd.Sender.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid sender format, %v", cmd.Sender.String())
	} else if target, err := cmd.Contract.Encode(enc); err != nil {
		return errors.Wrapf(err, "invalid contract address format, %v", cmd.Contract.String())
	} else {
		cmd.sender = sender
		cmd.target = target
	}

	return nil
}

func (cmd *UpdateHandlerCommand) createOperation() (base.Operation, error) { // nolint:dupl
	handlers := make([]base.Address, len(cmd.Handlers))
	for i := range cmd.Handlers {
		ad, err := base.DecodeAddress(cmd.Handlers[i].String(), enc)
		if err != nil {
			return nil, err
		}

		handlers[i] = ad
	}

	fact := extension.NewUpdateHandlerFact([]byte(cmd.Token), cmd.sender, cmd.target, handlers, cmd.Currency.CID)

	op, err := extension.NewUpdateHandler(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create updateHandler operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "create updateHandler operation")
	}

	return op, nil
}
