package contract

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/pkg/errors"
)

func proposalBlockTime(proposal *base.ProposalSignFact) (int64, error) {
	if proposal == nil || *proposal == nil || (*proposal).ProposalFact() == nil {
		return 0, errors.Errorf("proposal is required for contract write block time")
	}

	// Contracts observe inclusion time in Unix seconds from the canonical
	// proposal fact, never the local wall clock of an executing node.
	return (*proposal).ProposalFact().ProposedAt().Unix(), nil
}
