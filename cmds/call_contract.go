package cmds

import (
	"context"
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/operation/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type CallContractCommand struct {
	BaseCommand
	OperationFlags
	Sender   AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract AddressFlag    `arg:"" name:"contract" help:"contract account to register policy" required:"true"`
	CallData string         `name:"calldata" help:"call data" required:"true"`
	Currency CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	sender   base.Address
	contract base.Address
	callData map[string]string
}

func (cmd *CallContractCommand) Run(pctx context.Context) error {
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

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

func (cmd *CallContractCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if a, err := cmd.Sender.Encode(cmd.Encoders.JSON()); err != nil {
		return errors.Wrapf(err, "invalid sender format; %q", cmd.Sender)
	} else {
		cmd.sender = a
	}

	if a, err := cmd.Contract.Encode(cmd.Encoders.JSON()); err != nil {
		return errors.Wrapf(err, "invalid contract format; %q", cmd.Contract)
	} else {
		cmd.contract = a
	}

	cmd.callData = make(map[string]string)
	json.Unmarshal([]byte(cmd.CallData), &cmd.callData)

	return nil
}

func (cmd *CallContractCommand) createOperation() (base.Operation, error) {
	e := util.StringError("failed to create register-model operation")

	fact := contract.NewCallContractFact([]byte(cmd.Token), cmd.sender, cmd.contract, cmd.callData, cmd.Currency.CID)

	op, err := contract.NewCallContract(fact)
	if err != nil {
		return nil, e.Wrap(err)
	}
	err = op.Sign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, e.Wrap(err)
	}

	return op, nil
}
