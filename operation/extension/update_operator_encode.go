package extension

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *UpdateOperatorFact) unpack(enc encoder.Encoder, sd, ct string, oprs []string, cid string) error {
	e := util.StringError("unmarshal UpdateOperatorFact")

	switch ad, err := base.DecodeAddress(sd, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.sender = ad
	}

	switch ad, err := base.DecodeAddress(ct, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.contract = ad
	}

	operators := make([]base.Address, len(oprs))
	for i := range oprs {
		switch ad, err := base.DecodeAddress(oprs[i], enc); {
		case err != nil:
			return e.Wrap(err)
		default:
			operators[i] = ad
		}
	}
	fact.operators = operators

	fact.currency = types.CurrencyID(cid)

	return nil
}
