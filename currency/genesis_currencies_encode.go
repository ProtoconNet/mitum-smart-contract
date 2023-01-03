package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
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

	var keys AccountKeys
	hinter, err := enc.Decode(bks)
	if err != nil {
		return e(err, "")
	} else if k, ok := hinter.(AccountKeys); !ok {
		return util.ErrWrongType.Errorf("expected AccountKeys, not %T", hinter)
	} else {
		keys = k
	}

	fact.keys = keys

	hcs, err := enc.DecodeSlice(bcs)
	if err != nil {
		return e(err, "")
	}

	cs := make([]CurrencyDesign, len(hcs))
	for i := range hcs {
		j, ok := hcs[i].(CurrencyDesign)
		if !ok {
			return util.ErrWrongType.Errorf("expected CurrencyDesign, not %T", hcs[i])
		}

		cs[i] = j
	}
	fact.cs = cs

	return nil
}
