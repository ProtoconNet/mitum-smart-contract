package cmds

import (
	"context"

	"github.com/ProtoconNet/mitum-currency/v3/operation/currency"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type CurrencyFixedFeeerFlags struct {
	Receiver AddressFlag `name:"receiver" help:"fee receiver account address"`
	Amount   BigFlag     `name:"amount" help:"fee amount"`
	feeer    types.Feeer
}

func (fl *CurrencyFixedFeeerFlags) IsValid([]byte) error {
	if len(fl.Receiver.String()) < 1 {
		return nil
	}

	var receiver base.Address
	if a, err := fl.Receiver.Encode(enc); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver format, %v: %v", fl.Receiver.String(), err)
	} else if err := a.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver address, %v: %v", fl.Receiver.String(), err)
	} else {
		receiver = a
	}

	fl.feeer = types.NewFixedFeeer(receiver, fl.Amount.Big)
	return fl.feeer.IsValid(nil)
}

type CurrencyRatioFeeerFlags struct {
	Receiver AddressFlag `name:"receiver" help:"fee receiver account address"`
	Ratio    float64     `name:"ratio" help:"fee ratio, multifly by operation amount"`
	Min      BigFlag     `name:"min" help:"minimum fee"`
	Max      BigFlag     `name:"max" help:"maximum fee"`
	feeer    types.Feeer
}

func (fl *CurrencyRatioFeeerFlags) IsValid([]byte) error {
	if len(fl.Receiver.String()) < 1 {
		return nil
	}

	var receiver base.Address
	if a, err := fl.Receiver.Encode(enc); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver format, %v: %v", fl.Receiver.String(), err)
	} else if err := a.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid receiver address, %v: %v", fl.Receiver.String(), err)
	} else {
		receiver = a
	}

	fl.feeer = types.NewRatioFeeer(receiver, fl.Ratio, fl.Min.Big, fl.Max.Big)
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
	currencyDesign          types.CurrencyDesign
}

func (fl *CurrencyDesignFlags) IsValid([]byte) error {
	if err := fl.CurrencyPolicyFlags.IsValid(nil); err != nil {
		return err
	} else if err := fl.CurrencyFixedFeeerFlags.IsValid(nil); err != nil {
		return err
	} else if err := fl.CurrencyRatioFeeerFlags.IsValid(nil); err != nil {
		return err
	}

	var feeer types.Feeer
	switch t := fl.FeeerString; t {
	case types.FeeerNil, "":
		feeer = types.NewNilFeeer()
	case types.FeeerFixed:
		feeer = fl.CurrencyFixedFeeerFlags.feeer
	case types.FeeerRatio:
		feeer = fl.CurrencyRatioFeeerFlags.feeer
	default:
		return util.ErrInvalid.Errorf("unknown feeer type, %v", t)
	}

	if feeer == nil {
		return util.ErrInvalid.Errorf("empty feeer flags")
	} else if err := feeer.IsValid(nil); err != nil {
		return err
	}

	po := types.NewCurrencyPolicy(fl.CurrencyPolicyFlags.NewAccountMinBalance.Big, feeer)
	if err := po.IsValid(nil); err != nil {
		return err
	}

	var genesisAccount base.Address
	if a, err := fl.GenesisAccount.Encode(enc); err != nil {
		return util.ErrInvalid.Errorf("invalid genesis-account format, %q: %v", fl.GenesisAccount.String(), err)
	} else if err := a.IsValid(nil); err != nil {
		return util.ErrInvalid.Errorf("invalid genesis-account address, %q: %v", fl.GenesisAccount.String(), err)
	} else {
		genesisAccount = a
	}

	am := types.NewAmount(fl.GenesisAmount.Big, fl.Currency.CID)
	if err := am.IsValid(nil); err != nil {
		return err
	}

	fl.currencyDesign = types.NewCurrencyDesign(am, genesisAccount, po)
	return fl.currencyDesign.IsValid(nil)
}

type RegisterCurrencyCommand struct {
	BaseCommand
	OperationFlags
	CurrencyDesignFlags
	Node AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node base.Address
}

func (cmd *RegisterCurrencyCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "failed to create register-currency operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid register-currency operation")
	} else {
		cmd.Log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *RegisterCurrencyCommand) parseFlags() error {
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

	cmd.Log.Debug().Interface("currency-design", cmd.CurrencyDesignFlags.currencyDesign).Msg("currency design loaded")

	return nil
}

func (cmd *RegisterCurrencyCommand) createOperation() (currency.RegisterCurrency, error) {
	fact := currency.NewRegisterCurrencyFact([]byte(cmd.Token), cmd.currencyDesign)

	op, err := currency.NewRegisterCurrency(fact, "")
	if err != nil {
		return currency.RegisterCurrency{}, err
	}

	err = op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node)
	if err != nil {
		return currency.RegisterCurrency{}, errors.Wrap(err, "failed to create register-currency operation")
	}

	return op, nil
}
