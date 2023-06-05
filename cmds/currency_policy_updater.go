package cmds

import (
	"context"
	base3 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/pkg/errors"

	"github.com/ProtoconNet/mitum2/base"
)

type CurrencyPolicyUpdaterCommand struct {
	baseCommand
	OperationFlags
	Currency                CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	CurrencyPolicyFlags     `prefix:"policy-" help:"currency policy" required:"true"`
	FeeerString             string `name:"feeer" help:"feeer type, {nil, fixed, ratio}" required:"true"`
	CurrencyFixedFeeerFlags `prefix:"feeer-fixed-" help:"fixed feeer"`
	CurrencyRatioFeeerFlags `prefix:"feeer-ratio-" help:"ratio feeer"`
	Node                    AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node                    base.Address
	po                      base3.CurrencyPolicy
}

func NewCurrencyPolicyUpdaterCommand() CurrencyPolicyUpdaterCommand {
	cmd := NewbaseCommand()
	return CurrencyPolicyUpdaterCommand{
		baseCommand: *cmd,
	}
}

func (cmd *CurrencyPolicyUpdaterCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	encs = cmd.encs
	enc = cmd.enc

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	var op base.Operation
	if i, err := cmd.createOperation(); err != nil {
		return errors.Wrap(err, "failed to create currency-policy-updater operation")
	} else if err := i.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return errors.Wrap(err, "invalid currency-policy-updater operation")
	} else {
		cmd.log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *CurrencyPolicyUpdaterCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	} else if err := cmd.CurrencyPolicyFlags.IsValid(nil); err != nil {
		return err
	}

	if err := cmd.CurrencyFixedFeeerFlags.IsValid(nil); err != nil {
		return err
	} else if err := cmd.CurrencyRatioFeeerFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Node.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid node format, %q", cmd.Node.String())
	}
	cmd.node = a

	var feeer base3.Feeer
	switch t := cmd.FeeerString; t {
	case base3.FeeerNil, "":
		feeer = base3.NewNilFeeer()
	case base3.FeeerFixed:
		feeer = cmd.CurrencyFixedFeeerFlags.feeer
	case base3.FeeerRatio:
		feeer = cmd.CurrencyRatioFeeerFlags.feeer
	default:
		return errors.Errorf("unknown feeer type, %q", t)
	}

	if feeer == nil {
		return errors.Errorf("empty feeer flags")
	} else if err := feeer.IsValid(nil); err != nil {
		return err
	}

	cmd.po = base3.NewCurrencyPolicy(cmd.CurrencyPolicyFlags.NewAccountMinBalance.Big, feeer)
	if err := cmd.po.IsValid(nil); err != nil {
		return err
	}

	cmd.log.Debug().Interface("currency-policy", cmd.po).Msg("currency policy loaded")

	return nil
}

func (cmd *CurrencyPolicyUpdaterCommand) createOperation() (currency.CurrencyPolicyUpdater, error) {
	fact := currency.NewCurrencyPolicyUpdaterFact([]byte(cmd.Token), cmd.Currency.CID, cmd.po)

	op, err := currency.NewCurrencyPolicyUpdater(fact, "")
	if err != nil {
		return currency.CurrencyPolicyUpdater{}, err
	}

	err = op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node)
	if err != nil {
		return currency.CurrencyPolicyUpdater{}, errors.Wrap(err, "failed to create currency-policy-updater operation")
	}

	return op, nil
}
