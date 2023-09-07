package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (de *CurrencyDesign) unpack(enc encoder.Encoder, ht hint.Hint, bam []byte, ga string, bpo []byte, ag string) error {
	e := util.StringError("unmarshal CurrencyDesign")

	de.BaseHinter = hint.NewBaseHinter(ht)

	var am Amount
	if err := encoder.Decode(enc, bam, &am); err != nil {
		return e.WithMessage(err, "decode amount")
	}

	de.amount = am

	switch ad, err := base.DecodeAddress(ga, enc); {
	case err != nil:
		return e.WithMessage(err, "decode address")
	default:
		de.genesisAccount = ad
	}

	var policy CurrencyPolicy

	if err := encoder.Decode(enc, bpo, &policy); err != nil {
		return e.WithMessage(err, "decode currency policy")
	}

	de.policy = policy

	if big, err := common.NewBigFromString(ag); err != nil {
		return e.Wrap(err)
	} else {
		de.aggregate = big
	}

	return nil
}
