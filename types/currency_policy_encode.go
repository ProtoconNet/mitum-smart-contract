package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

func (po *CurrencyPolicy) unpack(enc encoder.Encoder, ht hint.Hint, mn string, bfe []byte) error {
	if big, err := common.NewBigFromString(mn); err != nil {
		return err
	} else {
		po.minBalance = big
	}

	po.BaseHinter = hint.NewBaseHinter(ht)
	var feeer Feeer
	err := encoder.Decode(enc, bfe, &feeer)
	if err != nil {
		return errors.Errorf("Decode feeer, %v", err)
	}
	po.feeer = feeer

	return nil
}
