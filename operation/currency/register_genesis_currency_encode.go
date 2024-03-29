package currency

import (
	"github.com/ProtoconNet/mitum-currency/v3/common"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/pkg/errors"
)

func (fact *RegisterGenesisCurrencyFact) unpack(
	enc encoder.Encoder,
	gk string,
	bks []byte,
	bcs []byte,
) error {
	switch pk, err := base.DecodePublickeyFromString(gk, enc); {
	case err != nil:
		return err
	default:
		fact.genesisNodeKey = pk
	}

	var keys types.AccountKeys
	hinter, err := enc.Decode(bks)
	if err != nil {
		return err
	} else if k, ok := hinter.(types.AccountKeys); !ok {
		return common.ErrTypeMismatch.Wrap(errors.Errorf("expected AccountKeys, not %T", hinter))
	} else {
		keys = k
	}

	fact.keys = keys

	hcs, err := enc.DecodeSlice(bcs)
	if err != nil {
		return err
	}

	cs := make([]types.CurrencyDesign, len(hcs))
	for i := range hcs {
		j, ok := hcs[i].(types.CurrencyDesign)
		if !ok {
			return common.ErrTypeMismatch.Wrap(errors.Errorf("expected CurrencyDesign, not %T", hcs[i]))
		}

		cs[i] = j
	}
	fact.cs = cs

	return nil
}
