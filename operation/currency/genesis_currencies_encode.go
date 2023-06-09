package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

func (fact *GenesisCurrenciesFact) unpack(
	enc encoder.Encoder,
	gk string,
	bks []byte,
	bcs []byte,
) error {
	e := util.StringErrorFunc("failed to unmarshal GenesisCurrenciesFact")

	switch pk, err := base.DecodePublickeyFromString(gk, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.genesisNodeKey = pk
	}

	var keys types.AccountKeys
	hinter, err := enc.Decode(bks)
	if err != nil {
		return e(err, "")
	} else if k, ok := hinter.(types.AccountKeys); !ok {
		return util.ErrWrongType.Errorf("expected AccountKeys, not %T", hinter)
	} else {
		keys = k
	}

	fact.keys = keys

	hcs, err := enc.DecodeSlice(bcs)
	if err != nil {
		return e(err, "")
	}

	cs := make([]types.CurrencyDesign, len(hcs))
	for i := range hcs {
		j, ok := hcs[i].(types.CurrencyDesign)
		if !ok {
			return util.ErrWrongType.Errorf("expected CurrencyDesign, not %T", hcs[i])
		}

		cs[i] = j
	}
	fact.cs = cs

	return nil
}
