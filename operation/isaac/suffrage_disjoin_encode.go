package isaacoperation

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *SuffrageDisjoinFact) unpack(
	enc encoder.Encoder,
	nd string,
	height base.Height,
) error {
	switch i, err := base.DecodeAddress(nd, enc); {
	case err != nil:
		return err
	default:
		fact.node = i
	}

	fact.start = height

	return nil
}
