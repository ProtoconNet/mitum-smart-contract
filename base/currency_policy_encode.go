package base

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

func (po *CurrencyPolicy) unpack(enc encoder.Encoder, ht hint.Hint, mn string, bfe []byte) error {
	e := util.StringErrorFunc("failed to unmarshal CurrencyPolicy")

	if big, err := NewBigFromString(mn); err != nil {
		return e(err, "")
	} else {
		po.newAccountMinBalance = big
	}

	po.BaseHinter = hint.NewBaseHinter(ht)
	var feeer Feeer
	err := encoder.Decode(enc, bfe, &feeer)
	if err != nil {
		return e(err, "failed to decode feeer")
	}
	po.feeer = feeer

	return nil
}
