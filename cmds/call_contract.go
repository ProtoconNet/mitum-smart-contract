package cmds

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	ccmds "github.com/ProtoconNet/mitum-currency/v3/cmds"
	ctypes "github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum-smart-contract/operation/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type CallContractCommand struct {
	BaseCommand
	OperationFlags
	Sender    ccmds.AddressFlag    `arg:"" name:"sender" help:"sender address" required:"true"`
	Contract  ccmds.AddressFlag    `arg:"" name:"contract" help:"contract account to register policy" required:"true"`
	CallData  *string              `name:"calldata" help:"legacy single call data JSON object"`
	Items     *string              `name:"items" help:"batch call items JSON array"`
	ItemsFile *string              `name:"items-file" help:"path to batch call items JSON array file" type:"filepath"`
	Currency  ccmds.CurrencyIDFlag `arg:"" name:"currency" help:"currency id" required:"true"`
	sender    base.Address
	contract  base.Address
	fact      contract.CallContractFact
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

	ccmds.PrettyPrint(cmd.Out, op)

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

	fact, err := newCallContractFactFromCLIInput(
		[]byte(cmd.Token),
		cmd.sender,
		cmd.contract,
		cmd.CallData,
		cmd.Items,
		cmd.ItemsFile,
		cmd.Currency.CID,
	)
	if err != nil {
		return err
	}
	cmd.fact = fact

	return nil
}

func (cmd *CallContractCommand) createOperation() (base.Operation, error) {
	e := util.StringError("failed to create call-contract operation")

	op, err := contract.NewCallContract(cmd.fact)
	if err != nil {
		return nil, e.Wrap(err)
	}
	err = op.Sign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, e.Wrap(err)
	}

	return op, nil
}

type callContractItemInput struct {
	Function string            `json:"function"`
	CallData map[string]string `json:"call_data"`
}

func newCallContractFactFromCLIInput(
	token []byte,
	sender, contractAddr base.Address,
	callDataJSON, itemsJSON, itemsFile *string,
	currency ctypes.CurrencyID,
) (contract.CallContractFact, error) {
	selected := 0
	for _, input := range []*string{callDataJSON, itemsJSON, itemsFile} {
		if input != nil {
			selected++
		}
	}
	if selected != 1 {
		return contract.CallContractFact{}, errors.Errorf(
			"exactly one of --calldata, --items, or --items-file is required",
		)
	}

	switch {
	case callDataJSON != nil:
		callData, err := decodeCallContractLegacyCallData(*callDataJSON)
		if err != nil {
			return contract.CallContractFact{}, err
		}

		return contract.NewCallContractFact(token, sender, contractAddr, callData, currency), nil
	case itemsJSON != nil:
		items, err := decodeCallContractItemsJSON("--items", []byte(*itemsJSON))
		if err != nil {
			return contract.CallContractFact{}, err
		}

		return contract.NewCallContractFactWithItems(token, sender, contractAddr, items, currency), nil
	default:
		items, err := readCallContractItemsFile(*itemsFile)
		if err != nil {
			return contract.CallContractFact{}, err
		}

		return contract.NewCallContractFactWithItems(token, sender, contractAddr, items, currency), nil
	}
}

func decodeCallContractLegacyCallData(input string) (map[string]string, error) {
	var callData map[string]string
	if err := json.Unmarshal([]byte(input), &callData); err != nil {
		return nil, errors.Wrap(err, "invalid --calldata JSON object")
	}

	return callData, nil
}

func readCallContractItemsFile(path string) ([]contract.CallContractItem, error) {
	if path == "" {
		return nil, errors.Errorf("--items-file path is empty")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read --items-file %q", path)
	}

	return decodeCallContractItemsJSON("--items-file", b)
}

func decodeCallContractItemsJSON(name string, b []byte) ([]contract.CallContractItem, error) {
	if strings.TrimSpace(string(b)) == "" {
		return nil, errors.Errorf("%s must be a JSON array", name)
	}
	if strings.TrimLeft(string(b), " \t\r\n")[0] != '[' {
		return nil, errors.Errorf("%s must be a JSON array", name)
	}

	var inputs []callContractItemInput
	if err := json.Unmarshal(b, &inputs); err != nil {
		return nil, errors.Wrapf(err, "invalid %s JSON array", name)
	}

	items := make([]contract.CallContractItem, len(inputs))
	for i := range inputs {
		items[i] = contract.NewCallContractItem(inputs[i].Function, inputs[i].CallData)
	}

	return items, nil
}
