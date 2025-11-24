package cmds

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/ProtoconNet/mitum-currency/v3/operation/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type RegisterContractCommand struct {
	BaseCommand
	OperationFlags
	Sender   AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract AddressFlag    `arg:"" name:"contract" help:"contract account to register policy" required:"true"`
	Code     string         `arg:"" name:"contract-code" help:"contract code" type:"filepath"`
	CallData string         `name:"calldata" help:"call data" required:"true"`
	Currency CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	sender   base.Address
	contract base.Address
	code     string
	callData map[string]string
}

func (cmd *RegisterContractCommand) Run(pctx context.Context) error {
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

func (cmd *RegisterContractCommand) parseFlags() error {
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

	f, err := os.Open(cmd.Code)
	if err != nil {
		return errors.Wrapf(err, "invalid contract code path; %q", cmd.Code)
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	cmd.code = string(bytes)
	cmd.callData = make(map[string]string)
	json.Unmarshal([]byte(cmd.CallData), &cmd.callData)

	return nil
}

func (cmd *RegisterContractCommand) createOperation() (base.Operation, error) {
	e := util.StringError("failed to create register-model operation")

	fact := contract.NewRegisterContractFact([]byte(cmd.Token), cmd.sender, cmd.contract, cmd.code, cmd.callData, cmd.Currency.CID)

	op, err := contract.NewRegisterContract(fact)
	if err != nil {
		return nil, e.Wrap(err)
	}
	err = op.Sign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, e.Wrap(err)
	}

	return op, nil
}
