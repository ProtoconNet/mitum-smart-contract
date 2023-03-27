package cmds

import (
	"context"

	isaacoperation "github.com/ProtoconNet/mitum-currency/v2/isaac"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type SuffrageJoinCommand struct {
	baseCommand
	OperationFlags
	Node  AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	Start base.Height `arg:"" name:"height" help:"block height" required:"true"`
	node  base.Address
}

func NewSuffrageJoinCommand() SuffrageJoinCommand {
	cmd := NewbaseCommand()
	return SuffrageJoinCommand{
		baseCommand: *cmd,
	}
}

func (cmd *SuffrageJoinCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "failed to create suffrage-join operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid suffrage-join operation")
	} else {
		cmd.log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *SuffrageJoinCommand) parseFlags() error {
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

func (cmd *SuffrageJoinCommand) createOperation() (isaacoperation.SuffrageJoin, error) {
	fact := isaacoperation.NewSuffrageJoinFact([]byte(cmd.Token), cmd.node, cmd.Start)

	op := isaacoperation.NewSuffrageJoin(fact)
	if err := op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node); err != nil {
		return isaacoperation.SuffrageJoin{}, errors.Wrap(err, "failed to create suffrage-join operation")
	}

	return op, nil
}
