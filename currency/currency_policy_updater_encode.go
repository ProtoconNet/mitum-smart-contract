package currency

import (
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *CurrencyPolicyUpdaterFact) unpack(enc encoder.Encoder, cid string, bpo []byte) error {
	e := util.StringErrorFunc("failed to unmarshal CurrencyPolicyUpdaterFact")

	if hinter, err := enc.Decode(bpo); err != nil {
		return e(err, "")
	} else if po, ok := hinter.(CurrencyPolicy); !ok {
		return util.ErrWrongType.Errorf("expected CurrencyPolicy, not %T", hinter)
	} else {
		fact.policy = po
	}

	fact.currency = CurrencyID(cid)

	return nil
}
