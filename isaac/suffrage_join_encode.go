package isaacoperation

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *SuffrageJoinFact) unpack(
	enc encoder.Encoder,
	candidate string,
	height base.Height,
) error {
	e := util.StringErrorFunc("failed to unmarshal SuffrageJoinFact")

	switch i, err := base.DecodeAddress(candidate, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.candidate = i
	}

	fact.start = height

	return nil
}
