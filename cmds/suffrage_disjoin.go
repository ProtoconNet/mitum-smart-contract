package cmds

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	isaac "github.com/spikeekips/mitum/isaac/operation"
)

type SuffrageDisjoinCommand struct {
	baseCommand
	OperationFlags
	Node  AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	Start base.Height `arg:"" name:"height" help:"block height" required:"true"`
	node  base.Address
}

func NewSuffrageDisjoinCommand() SuffrageDisjoinCommand {
	cmd := NewbaseCommand()
	return SuffrageDisjoinCommand{
		baseCommand: *cmd,
	}
}

func (cmd *SuffrageDisjoinCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "failed to create suffrage-disjoin operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid suffrage-disjoin operation")
	} else {
		cmd.log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *SuffrageDisjoinCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Node.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid node format, %q", cmd.Node.String())
	}
	cmd.node = a

	return nil
}

func (cmd *SuffrageDisjoinCommand) createOperation() (isaac.SuffrageDisjoin, error) {
	fact := isaac.NewSuffrageDisjoinFact([]byte(cmd.Token), cmd.node, cmd.Start)

	op := isaac.NewSuffrageDisjoin(fact)
	if err := op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node); err != nil {
		return isaac.SuffrageDisjoin{}, errors.Wrap(err, "failed to create suffrage-disjoin operation")
	}

	return op, nil
}
