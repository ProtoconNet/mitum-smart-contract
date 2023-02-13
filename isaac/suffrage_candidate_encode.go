package isaacoperation

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *SuffrageCandidateFact) unpack(
	enc encoder.Encoder,
	sd string,
	pk string,
) error {
	e := util.StringErrorFunc("failed to unmarshal SuffrageCandidateFact")

	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.address = ad
	}

	switch p, err := base.DecodePublickeyFromString(pk, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.publickey = p
	}

	return nil
}
