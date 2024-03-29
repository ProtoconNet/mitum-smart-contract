package isaacoperation

import (
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *SuffrageCandidateFact) unpack(
	enc encoder.Encoder,
	sd string,
	pk string,
) error {
	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return err
	default:
		fact.address = ad
	}

	switch p, err := base.DecodePublickeyFromString(pk, enc); {
	case err != nil:
		return err
	default:
		fact.publickey = p
	}

	return nil
}
