package types

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (p *NetworkPolicy) unpack(
	enc encoder.Encoder,
	suffrageCandidateLimiterRule []byte,
	maxOperationsInProposal uint64,
	suffrageCandidateLifespan base.Height,
	maxSuffrageSize uint64,
	suffrageExpelLifespan base.Height,
	emptyProposalNoBlock bool,
) error {
	e := util.StringError("unmarshal NetworkPolicy")

	if err := encoder.Decode(enc, suffrageCandidateLimiterRule, &p.suffrageCandidateLimiterRule); err != nil {
		return e.Wrap(err)
	}

	p.maxOperationsInProposal = maxOperationsInProposal
	p.suffrageCandidateLifespan = suffrageCandidateLifespan
	p.maxSuffrageSize = maxSuffrageSize
	p.suffrageExpelLifespan = suffrageExpelLifespan
	p.emptyProposalNoBlock = emptyProposalNoBlock

	return nil
}
