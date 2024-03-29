package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (de *CurrencyDesign) unpack(enc encoder.Encoder, ht hint.Hint, bam []byte, ga string, bpo []byte, ag string) error {
	de.BaseHinter = hint.NewBaseHinter(ht)

	var am Amount
	if err := encoder.Decode(enc, bam, &am); err != nil {
		return errors.Errorf("Decode amount, %v", err)
	}

	de.amount = am

	switch ad, err := base.DecodeAddress(ga, enc); {
	case err != nil:
		return errors.Errorf("Decode address, %v", err)
	default:
		de.genesisAccount = ad
	}

	var policy CurrencyPolicy

	if err := encoder.Decode(enc, bpo, &policy); err != nil {
		return errors.Errorf("Decode currency policy, %v", err)
	}

	de.policy = policy

	if big, err := common.NewBigFromString(ag); err != nil {
		return err
	} else {
		de.aggregate = big
	}

	return nil
}
