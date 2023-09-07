package types

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (po *CurrencyPolicy) unpack(enc encoder.Encoder, ht hint.Hint, mn string, bfe []byte) error {
	e := util.StringError("unmarshal CurrencyPolicy")

	if big, err := common.NewBigFromString(mn); err != nil {
		return e.Wrap(err)
	} else {
		po.newAccountMinBalance = big
	}

	po.BaseHinter = hint.NewBaseHinter(ht)
	var feeer Feeer
	err := encoder.Decode(enc, bfe, &feeer)
	if err != nil {
		return e.WithMessage(err, "decode feeer")
	}
	po.feeer = feeer

	return nil
}
