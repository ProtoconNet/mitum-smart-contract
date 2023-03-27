package cmds

import (
	"context"

	isaacoperation "github.com/ProtoconNet/mitum-currency/v2/isaac"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type SuffrageCandidateCommand struct {
	baseCommand
	OperationFlags
	Node      AddressFlag   `arg:"" name:"node" help:"node address" required:"true"`
	PublicKey PublickeyFlag `arg:"" name:"public-key" help:"public key" required:"true"`
	node      base.Address
}

func NewSuffrageCandidateCommand() SuffrageCandidateCommand {
	cmd := NewbaseCommand()
	return SuffrageCandidateCommand{
		baseCommand: *cmd,
	}
}

func (cmd *SuffrageCandidateCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "failed to create suffrage-candidate operation")
	} else if err := i.IsValid([]byte(cmd.OperationFlags.NetworkID)); err != nil {
		return errors.Wrap(err, "invalid suffrage-candidate operation")
	} else {
		cmd.log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *SuffrageCandidateCommand) parseFlags() error {
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

func (cmd *SuffrageCandidateCommand) createOperation() (isaacoperation.SuffrageCandidate, error) {
	fact := isaacoperation.NewSuffrageCandidateFact([]byte(cmd.Token), cmd.node, cmd.PublicKey.Publickey)

	op := isaacoperation.NewSuffrageCandidate(fact)
	if err := op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node); err != nil {
		return isaacoperation.SuffrageCandidate{}, errors.Wrap(err, "failed to create suffrage-candidate operation")
	}

	return op, nil
}
