package isaacoperation

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *SuffrageDisjoinFact) unpack(
	enc encoder.Encoder,
	nd string,
	height base.Height,
) error {
	e := util.StringErrorFunc("failed to unmarshal SuffrageDisjoinFact")

	switch i, err := base.DecodeAddress(nd, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.node = i
	}

	fact.start = height

	return nil
}
