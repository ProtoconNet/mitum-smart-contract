package currency

import (
	base2 "github.com/ProtoconNet/mitum-currency/v3/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CurrencyPolicyUpdaterFact) unpack(enc encoder.Encoder, cid string, bpo []byte) error {
	e := util.StringErrorFunc("failed to unmarshal CurrencyPolicyUpdaterFact")

	if hinter, err := enc.Decode(bpo); err != nil {
		return e(err, "")
	} else if po, ok := hinter.(base2.CurrencyPolicy); !ok {
		return util.ErrWrongType.Errorf("expected CurrencyPolicy, not %T", hinter)
	} else {
		fact.policy = po
	}

	fact.currency = base2.CurrencyID(cid)

	return nil
}
