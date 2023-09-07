package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *RegisterGenesisCurrencyFact) unpack(
	enc encoder.Encoder,
	gk string,
	bks []byte,
	bcs []byte,
) error {
	e := util.StringError("unmarshal RegisterGenesisCurrencyFact")

	switch pk, err := base.DecodePublickeyFromString(gk, enc); {
	case err != nil:
		return e.Wrap(err)
	default:
		fact.genesisNodeKey = pk
	}

	var keys types.AccountKeys
	hinter, err := enc.Decode(bks)
	if err != nil {
		return e.Wrap(err)
	} else if k, ok := hinter.(types.AccountKeys); !ok {
		return errors.Errorf("expected AccountKeys, not %T", hinter)
	} else {
		keys = k
	}

	fact.keys = keys

	hcs, err := enc.DecodeSlice(bcs)
	if err != nil {
		return e.Wrap(err)
	}

	cs := make([]types.CurrencyDesign, len(hcs))
	for i := range hcs {
		j, ok := hcs[i].(types.CurrencyDesign)
		if !ok {
			return errors.Errorf("expected CurrencyDesign, not %T", hcs[i])
		}

		cs[i] = j
	}
	fact.cs = cs

	return nil
}
