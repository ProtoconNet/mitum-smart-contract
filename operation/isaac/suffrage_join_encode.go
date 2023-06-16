package isaacoperation

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *SuffrageJoinFact) unpack(
	enc encoder.Encoder,
	candidate string,
	height base.Height,
) error {
	e := util.StringError("failed to unmarshal SuffrageJoinFact")

	switch i, err := base.DecodeAddress(candidate, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.candidate = i
	}

	fact.start = height

	return nil
}
