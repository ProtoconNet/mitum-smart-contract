package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (de *CurrencyDesign) unpack(enc encoder.Encoder, ht hint.Hint, isp, cr, dc, ga string, bpo []byte, ts string) error {
	de.BaseHinter = hint.NewBaseHinter(ht)

	if initialSupply, err := common.NewBigFromString(isp); err != nil {
		return err
	} else {
		de.initialSupply = initialSupply
	}

	currencyID := CurrencyID(cr)
	if err := currencyID.IsValid(nil); err != nil {
		return err
	}
	de.currency = currencyID

	if decimal, err := common.NewBigFromString(dc); err != nil {
		return err
	} else {
		de.decimal = decimal
	}

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

	if big, err := common.NewBigFromString(ts); err != nil {
		return err
	} else {
		de.totalSupply = big
	}

	return nil
}
