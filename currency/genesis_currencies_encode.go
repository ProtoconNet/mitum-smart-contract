package currency

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
)

func (fact *GenesisCurrenciesFact) unpack(
	enc encoder.Encoder,
	ufact GenesisCurrenciesFactJSONUnMarshaler,
) error {
	fact.BaseFact.SetJSONUnmarshaler(ufact.BaseFactJSONUnmarshaler)

	switch pk, err := base.DecodePublickeyFromString(ufact.GK, enc); {
	case err != nil:
		return err
	default:
		fact.genesisNodeKey = pk
	}

	var keys AccountKeys
	hinter, err := enc.Decode(ufact.KS)
	if err != nil {
		return err
	} else if k, ok := hinter.(AccountKeys); !ok {
		return errors.Errorf("not Keys: %T", hinter)
	} else {
		keys = k
	}

	fact.keys = keys

	hcs, err := enc.DecodeSlice(ufact.CS)
	if err != nil {
		return err
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
