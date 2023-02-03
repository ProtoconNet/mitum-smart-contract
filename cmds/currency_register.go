package cmds

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum-currency/currency"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
)

type CurrencyFixedFeeerFlags struct {
	Receiver AddressFlag `name:"receiver" help:"fee receiver account address"`
	Amount   BigFlag     `name:"amount" help:"fee amount"`
	feeer    currency.Feeer
}

func (fl *CurrencyFixedFeeerFlags) IsValid([]byte) error {
	if len(fl.Receiver.String()) < 1 {
		return nil
	}

	var receiver base.Address
	if a, err := fl.Receiver.Encode(enc); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver format, %q: %w", fl.Receiver.String(), err)
	} else if err := a.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver address, %q: %w", fl.Receiver.String(), err)
	} else {
		receiver = a
	}

	fl.feeer = currency.NewFixedFeeer(receiver, fl.Amount.Big)
	return fl.feeer.IsValid(nil)
}

type CurrencyRatioFeeerFlags struct {
	Receiver AddressFlag `name:"receiver" help:"fee receiver account address"`
	Ratio    float64     `name:"ratio" help:"fee ratio, multifly by operation amount"`
	Min      BigFlag     `name:"min" help:"minimum fee"`
	Max      BigFlag     `name:"max" help:"maximum fee"`
	feeer    currency.Feeer
}

func (fl *CurrencyRatioFeeerFlags) IsValid([]byte) error {
	if len(fl.Receiver.String()) < 1 {
		return nil
	}

	var receiver base.Address
	if a, err := fl.Receiver.Encode(enc); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver format, %q: %w", fl.Receiver.String(), err)
	} else if err := a.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver address, %q: %w", fl.Receiver.String(), err)
	} else {
		receiver = a
	}

	fl.feeer = currency.NewRatioFeeer(receiver, fl.Ratio, fl.Min.Big, fl.Max.Big)
	return fl.feeer.IsValid(nil)
}

type CurrencyPolicyFlags struct {
	NewAccountMinBalance BigFlag `name:"new-account-min-balance" help:"minimum balance for new account"` // nolint lll
}

func (*CurrencyPolicyFlags) IsValid([]byte) error {
	return nil
}

type CurrencyDesignFlags struct {
	Currency                CurrencyIDFlag `arg:"" name:"currency-id" help:"currency id" required:"true"`
	GenesisAmount           BigFlag        `arg:"" name:"genesis-amount" help:"genesis amount" required:"true"`
	GenesisAccount          AddressFlag    `arg:"" name:"genesis-account" help:"genesis-account address for genesis balance" required:"true"` // nolint lll
	CurrencyPolicyFlags     `prefix:"policy-" help:"currency policy" required:"true"`
	FeeerString             string `name:"feeer" help:"feeer type, {nil, fixed, ratio}" required:"true"`
	CurrencyFixedFeeerFlags `prefix:"feeer-fixed-" help:"fixed feeer"`
	CurrencyRatioFeeerFlags `prefix:"feeer-ratio-" help:"ratio feeer"`
	currencyDesign          currency.CurrencyDesign
}

func (fl *CurrencyDesignFlags) IsValid([]byte) error {
	if err := fl.CurrencyPolicyFlags.IsValid(nil); err != nil {
		return err
	} else if err := fl.CurrencyFixedFeeerFlags.IsValid(nil); err != nil {
		return err
	} else if err := fl.CurrencyRatioFeeerFlags.IsValid(nil); err != nil {
		return err
	}

	var feeer currency.Feeer
	switch t := fl.FeeerString; t {
	case currency.FeeerNil, "":
		feeer = currency.NewNilFeeer()
	case currency.FeeerFixed:
		feeer = fl.CurrencyFixedFeeerFlags.feeer
	case currency.FeeerRatio:
		feeer = fl.CurrencyRatioFeeerFlags.feeer
	default:
		return util.ErrInvalid.Errorf("unknown feeer type, %q", t)
	}

	if feeer == nil {
		return util.ErrInvalid.Errorf("empty feeer flags")
	} else if err := feeer.IsValid(nil); err != nil {
		return err
	}

	po := currency.NewCurrencyPolicy(fl.CurrencyPolicyFlags.NewAccountMinBalance.Big, feeer)
	if err := po.IsValid(nil); err != nil {
		return err
	}

	var genesisAccount base.Address
	if a, err := fl.GenesisAccount.Encode(enc); err != nil {
		return util.ErrInvalid.Errorf("invalid genesis-account format, %q: %w", fl.GenesisAccount.String(), err)
	} else if err := a.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid genesis-account address, %q: %w", fl.GenesisAccount.String(), err)
	} else {
		genesisAccount = a
	}

	am := currency.NewAmount(fl.GenesisAmount.Big, fl.Currency.CID)
	if err := am.IsValid(nil); err != nil {
		return err
	}

	fl.currencyDesign = currency.NewCurrencyDesign(am, genesisAccount, po)
	return fl.currencyDesign.IsValid(nil)
}

type CurrencyRegisterCommand struct {
	baseCommand
	OperationFlags
	CurrencyDesignFlags
	Node AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node base.Address
}

func NewCurrencyRegisterCommand() CurrencyRegisterCommand {
	cmd := NewbaseCommand()
	return CurrencyRegisterCommand{
		baseCommand: *cmd,
	}
}

func (cmd *CurrencyRegisterCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "failed to create currency-register operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid currency-register operation")
	} else {
		cmd.log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	/*
		i, err := base.NewBaseSeal(
			cmd.OperationFlags.Privatekey,
			[]operation.Operation{op},
			[]byte(cmd.OperationFlags.NetworkID),
		)
		if err != nil {
			return errors.Wrap(err, "failed to create operation.Seal")
		}
		cmd.log.Debug().Interface("seal", i).Msg("seal loaded")
	*/

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *CurrencyRegisterCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	} else if err := cmd.CurrencyDesignFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Node.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid node format, %q", cmd.Node.String())
	}
	cmd.node = a

	cmd.log.Debug().Interface("currency-design", cmd.CurrencyDesignFlags.currencyDesign).Msg("currency design loaded")

	return nil
}

func (cmd *CurrencyRegisterCommand) createOperation() (currency.CurrencyRegister, error) {
	fact := currency.NewCurrencyRegisterFact([]byte(cmd.Token), cmd.currencyDesign)

	op, err := currency.NewCurrencyRegister(fact, "")
	if err != nil {
		return currency.CurrencyRegister{}, err
	}

	err = op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node)
	if err != nil {
		return currency.CurrencyRegister{}, errors.Wrap(err, "failed to create currency-register operation")
	}

	return op, nil
}
