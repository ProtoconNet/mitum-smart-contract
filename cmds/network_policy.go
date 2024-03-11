package cmds

import (
	"context"
	isaacoperation "github.com/ProtoconNet/mitum-currency/v3/operation/isaac"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

type NetworkPolicyCommand struct {
	BaseCommand
	OperationFlags
	suffrageCandidateLimit    uint64      `help:"limit for suffrage candidates" default:"${suffrage_candidate_limiter_limit}"` // nolint
	MaxOperationInProposal    uint64      `help:"max operation in proposal" default:"${max_operation_in_proposal}"`            // nolint
	SuffrageCandidateLifespan uint64      `help:"suffrage candidate lifespan" default:"${max_operation_in_proposal}"`          // nolint
	MaxSuffrageSize           uint64      `help:"max suffrage size" default:"${max_operation_in_proposal}"`                    // nolint
	SuffrageExpelLifespan     uint64      `help:"suffrage expel lifespan" default:"${max_operation_in_proposal}"`              // nolint
	EmptyProposalNoBlock      bool        `help:"empty proposal no block"`                                                     // nolint
	Node                      AddressFlag `arg:"" name:"node" help:"node address" required:"true"`
	node                      base.Address
	policy                    base.NetworkPolicy
}

func (cmd *NetworkPolicyCommand) Run(pctx context.Context) error { // nolint:dupl
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
		return errors.Wrap(err, "create suffrage-candidate operation")
	} else if err := i.IsValid(cmd.OperationFlags.NetworkID); err != nil {
		return errors.Wrap(err, "invalid suffrage-candidate operation")
	} else {
		cmd.Log.Debug().Interface("operation", i).Msg("operation loaded")

		op = i
	}

	PrettyPrint(cmd.Out, op)

	return nil
}

func (cmd *NetworkPolicyCommand) parseFlags() error {
	if err := cmd.OperationFlags.IsValid(nil); err != nil {
		return err
	}

	a, err := cmd.Node.Encode(enc)
	if err != nil {
		return errors.Wrapf(err, "invalid node format, %v", cmd.Node.String())
	}
	cmd.node = a

	cmd.policy = types.NewNetworkPolicy(
		cmd.suffrageCandidateLimit,
		cmd.MaxOperationInProposal,
		base.Height(cmd.SuffrageCandidateLifespan),
		cmd.MaxSuffrageSize,
		base.Height(cmd.SuffrageExpelLifespan),
		cmd.EmptyProposalNoBlock,
	)

	return nil
}

func (cmd *NetworkPolicyCommand) createOperation() (isaacoperation.NetworkPolicy, error) {
	fact := isaacoperation.NewNetworkPolicyFact(
		[]byte(cmd.Token),
		cmd.policy,
	)

	op := isaacoperation.NewNetworkPolicy(fact)
	if err := op.NodeSign(cmd.Privatekey, cmd.NetworkID.NetworkID(), cmd.node); err != nil {
		return isaacoperation.NetworkPolicy{}, errors.Wrap(err, "create network-policy operation")
	}

	return op, nil
}
