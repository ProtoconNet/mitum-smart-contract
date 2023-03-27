package cmds

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtoconNet/mitum-currency/v2/currency"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/pkg/errors"
)

type SuffrageInflationItemFlag struct {
	s        string
	receiver base.Address
	amount   currency.Amount
}

func (v *SuffrageInflationItemFlag) String() string {
	return v.s
}

func (v *SuffrageInflationItemFlag) UnmarshalText(b []byte) error {
	v.s = string(b)

	l := strings.SplitN(string(b), ",", 3)
	if len(l) != 3 {
		return util.ErrInvalid.Errorf("invalid inflation amount, %q", string(b))
	}

	a, c := l[0], l[1]+","+l[2]

	af := &AddressFlag{}
	if err := af.UnmarshalText([]byte(a)); err != nil {
		return util.ErrInvalid.Errorf("invalid inflation receiver address: %w", err)
	}

	receiver, err := af.Encode(enc)
	if err != nil {
		return util.ErrInvalid.Errorf("invalid inflation receiver address: %w", err)
	}

	v.receiver = receiver

	cf := &CurrencyAmountFlag{}
	if err := cf.UnmarshalText([]byte(c)); err != nil {
		return util.ErrInvalid.Errorf("invalid inflation amount: %w", err)
	}
	v.amount = currency.NewAmount(cf.Big, cf.CID)

	return nil
}

func (v *SuffrageInflationItemFlag) IsValid([]byte) error {
	if err := util.CheckIsValiders(nil, false, v.receiver, v.amount); err != nil {
		return err
	}

	if !v.amount.Big().OverZero() {
		return util.ErrInvalid.Errorf("amount should be over zero")
	}

	return nil
}

type SuffrageInflationCommand struct {
	baseCommand
	OperationFlags
	Node  AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node  base.Address
	Items []SuffrageInflationItemFlag `arg:"" name:"inflation item" help:"ex: \"<receiver address>,<currency>,<amount>\""`
	items []currency.SuffrageInflationItem
}

func NewSuffrageInflationCommand() SuffrageInflationCommand {
	cmd := NewbaseCommand()
	return SuffrageInflationCommand{
		baseCommand: *cmd,
	}
}

func (cmd *SuffrageInflationCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "failed to create suffrage-inflation operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid suffrage-inflation operation")
	} else {
		cmd.log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *SuffrageInflationCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	if len(cmd.Items) < 1 {
		return fmt.Errorf("empty item flags")
	}

	items := make([]currency.SuffrageInflationItem, len(cmd.Items))
	for i := range cmd.Items {
		item := cmd.Items[i]
		if err := item.IsValid(nil); err != nil {
			return err
		}

		items[i] = currency.NewSuffrageInflationItem(item.receiver, item.amount)

		cmd.log.Debug().
			Stringer("amount", item.amount).
			Stringer("receiver", item.receiver).
			Msg("inflation item loaded")
	}
	cmd.items = items

	a, err := cmd.Node.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid node format, %q", cmd.Node.String())
	}
	cmd.node = a

	return nil
}

func (cmd *SuffrageInflationCommand) createOperation() (currency.SuffrageInflation, error) {
	fact := currency.NewSuffrageInflationFact([]byte(cmd.Token), cmd.items)

	op, err := currency.NewSuffrageInflation(fact)
	if err != nil {
		return currency.SuffrageInflation{}, err
	}

	err = op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node)
	if err != nil {
		return currency.SuffrageInflation{}, errors.Wrap(err, "failed to create suffrage-inflation operation")
	}

	return op, nil
}
