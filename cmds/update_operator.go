package cmds

import (
	"context"
	"github.com/ProtoconNet/mitum-currency/v3/operation/extension"
	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type UpdateOperatorCommand struct {
	BaseCommand
	OperationFlags
	Sender    AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract  AddressFlag    `arg:"" name:"contract" help:"target contract account address" required:"true"`
	Currency  CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	Operators []AddressFlag  `arg:"" name:"operators" help:"operators"`
	sender    base.Address
	target    base.Address
}

func (cmd *UpdateOperatorCommand) Run(pctx context.Context) error {
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

func (cmd *UpdateOperatorCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Operators) < 1 {
		return errors.Errorf("empty operators, must be given at least one")
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

func (cmd *UpdateOperatorCommand) createOperation() (base.Operation, error) { // nolint:dupl
	operators := make([]base.Address, len(cmd.Operators))
	for i := range cmd.Operators {
		ad, err := base.DecodeAddress(cmd.Operators[i].String(), enc)
		if err != nil {
			return nil, err
		}

		operators[i] = ad
	}

	fact := extension.NewUpdateOperatorFact([]byte(cmd.Token), cmd.sender, cmd.target, operators, cmd.Currency.CID)

	op, err := extension.NewUpdateOperator(fact)
	if err != nil {
		return nil, errors.Wrap(err, "create updateOperator operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "create updateOperator operation")
	}

	return op, nil
}
