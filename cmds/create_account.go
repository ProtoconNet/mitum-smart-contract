package cmds

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum-currency/v2/currency"
	"github.com/ProtoconNet/mitum2/base"
)

type CreateAccountCommand struct {
	baseCommand
	OperationFlags
	Sender      AddressFlag          `arg:"" name:"sender" help:"sender address" required:"true"`
	Threshold   uint                 `help:"threshold for keys (default: ${create_account_threshold})" default:"${create_account_threshold}"` // nolint
	Keys        []KeyFlag            `name:"key" help:"key for new account (ex: \"<public key>,<weight>\")" sep:"@"`
	Amounts     []CurrencyAmountFlag `arg:"" name:"currency-amount" help:"amount (ex: \"<currency>,<amount>\")"`
	AddressType string               `help:"address type for new account select mitum or ether" default:"mitum"`
	sender      base.Address
	keys        currency.BaseAccountKeys
}

func NewCreateAccountCommand() CreateAccountCommand {
	cmd := NewbaseCommand()
	return CreateAccountCommand{
		baseCommand: *cmd,
	}
}

func (cmd *CreateAccountCommand) Run(pctx context.Context) error { // nolint:dupl
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

	/*
		sl, err := LoadSealAndAddOperation(
			cmd.Seal.Bytes(),
			cmd.Privatekey,
			cmd.NetworkID.NetworkID(),
			op,
		)
		if err != nil {
			return err
		}
	*/
	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *CreateAccountCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Sender.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid sender format, %q", cmd.Sender.String())
	}
	cmd.sender = a

	if len(cmd.Keys) < 1 {
		return errors.Errorf("--key must be given at least one")
	}

	if len(cmd.Amounts) < 1 {
		return errors.Errorf("empty currency-amount, must be given at least one")
	}

	{
		ks := make([]currency.AccountKey, len(cmd.Keys))
		for i := range cmd.Keys {
			ks[i] = cmd.Keys[i].Key
		}

		if kys, err := currency.NewBaseAccountKeys(ks, cmd.Threshold); err != nil {
			return err
		} else if err := kys.IsValid(nil); err != nil {
			return err
		} else {
			cmd.keys = kys
		}
	}

	return nil
}

func (cmd *CreateAccountCommand) createOperation() (base.Operation, error) { // nolint:dupl}
	var items []currency.CreateAccountsItem

	ams := make([]currency.Amount, len(cmd.Amounts))
	for i := range cmd.Amounts {
		a := cmd.Amounts[i]
		am := currency.NewAmount(a.Big, a.CID)
		if err := am.IsValid(nil); err != nil {
			return nil, err
		}

		ams[i] = am
	}

	addrType := currency.AddressHint.Type()

	if cmd.AddressType == "ether" {
		addrType = currency.EthAddressHint.Type()
	}

	item := currency.NewCreateAccountsItemMultiAmounts(cmd.keys, ams, addrType)
	if err := item.IsValid(nil); err != nil {
		return nil, err
	}
	items = append(items, item)

	fact := currency.NewCreateAccountsFact([]byte(cmd.Token), cmd.sender, items)

	op, err := currency.NewCreateAccounts(fact)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create create-account operation")
	}
	err = op.HashSign(cmd.Privatekey, cmd.NetworkID.NetworkID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create create-account operation")
	}

	return op, nil
}

/*
func LoadSeal(b []byte, networkID base.NetworkID) (base.Seal, error) {
	if len(bytes.TrimSpace(b)) < 1 {
		return nil, errors.Errorf("empty input")
	}

	var sl base.Seal
	if err := encoder.Decode(b, enc, &sl); err != nil {
		return nil, err
	}

	if err := sl.IsValid(networkID); err != nil {
		return nil, errors.Wrap(err, "invalid seal")
	}

	return sl, nil
}

func LoadSealAndAddOperation(
	b []byte,
	privatekey key.Privatekey,
	networkID base.NetworkID,
	op operation.Operation,
) (operation.Seal, error) {
	if b == nil {
		bs, err := operation.NewBaseSeal(
			privatekey,
			[]operation.Operation{op},
			networkID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create operation.Seal")
		}
		return bs, nil
	}

	var sl operation.Seal
	if s, err := LoadSeal(b, networkID); err != nil {
		return nil, err
	} else if so, ok := s.(operation.Seal); !ok {
		return nil, errors.Errorf("seal is not operation.Seal, %T", s)
	} else if _, ok := so.(operation.SealUpdater); !ok {
		return nil, errors.Errorf("seal is not operation.SealUpdater, %T", s)
	} else {
		sl = so
	}

	// NOTE add operation to existing seal
	sl = sl.(operation.SealUpdater).SetOperations([]operation.Operation{op}).(operation.Seal)

	s, err := SignSeal(sl, privatekey, networkID)
	if err != nil {
		return nil, err
	}
	sl = s.(operation.Seal)

	return sl, nil
}
*/
