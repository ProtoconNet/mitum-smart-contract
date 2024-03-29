package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (am *Amount) unpack(enc encoder.Encoder, cid string, big string) error {
	am.cid = CurrencyID(cid)

	if b, err := common.NewBigFromString(big); err != nil {
		return err
	} else {
		am.big = b
	}

	return nil
}
