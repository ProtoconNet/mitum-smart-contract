package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type UpdateCurrencyCommand struct {
	BaseCommand
	OperationFlags
	Currency                CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	CurrencyPolicyFlags     `prefix:"policy-" help:"currency policy" required:"true"`
	FeeerString             string `name:"feeer" help:"feeer type, {nil, fixed, ratio}" required:"true"`
	CurrencyFixedFeeerFlags `prefix:"feeer-fixed-" help:"fixed feeer"`
	CurrencyRatioFeeerFlags `prefix:"feeer-ratio-" help:"ratio feeer"`
	Node                    AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node                    base.Address
	po                      types.CurrencyPolicy
}

func (cmd *UpdateCurrencyCommand) Run(pctx context.Context) error { // nolint:dupl
	if _, err := cmd.prepare(pctx); err != nil {
		return err
	}

	encs = cmd.Encoders
	enc = cmd.Encoder

	if err := cmd.parseFlags(); err != nil {
		return err
	}

	var op base.Operation
	if i, err := cmd.createOperation(); err != nil {
		return errors.Wrap(err, "failed to create update-currency operation")
	} else if err := i.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return errors.Wrap(err, "invalid update-currency operation")
	} else {
		cmd.Log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *UpdateCurrencyCommand) parseFlags() error {
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

	var feeer types.Feeer
	switch t := cmd.FeeerString; t {
	case types.FeeerNil, "":
		feeer = types.NewNilFeeer()
	case types.FeeerFixed:
		feeer = cmd.CurrencyFixedFeeerFlags.feeer
	case types.FeeerRatio:
		feeer = cmd.CurrencyRatioFeeerFlags.feeer
	default:
		return errors.Errorf("unknown feeer type, %q", t)
	}

	if feeer == nil {
		return errors.Errorf("empty feeer flags")
	} else if err := feeer.IsValid(nil); err != nil {
		return err
	}

	cmd.po = types.NewCurrencyPolicy(cmd.CurrencyPolicyFlags.NewAccountMinBalance.Big, feeer)
	if err := cmd.po.IsValid(nil); err != nil {
		return err
	}

	cmd.Log.Debug().Interface("currency-policy", cmd.po).Msg("currency policy loaded")

	return nil
}

func (cmd *UpdateCurrencyCommand) createOperation() (currency.UpdateCurrency, error) {
	fact := currency.NewUpdateCurrencyFact([]byte(cmd.Token), cmd.Currency.CID, cmd.po)

	op, err := currency.NewUpdateCurrency(fact, "")
	if err != nil {
		return currency.UpdateCurrency{}, err
	}

	err = op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node)
	if err != nil {
		return currency.UpdateCurrency{}, errors.Wrap(err, "failed to create update-currency operation")
	}

	return op, nil
}
