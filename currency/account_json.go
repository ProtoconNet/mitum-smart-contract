package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/valuehash"
)

type AccountJSONMarshaler struct {
	hint.BaseHinter
	H  util.Hash    `json:"hash"`
	AD base.Address `json:"address"`
	KS AccountKeys  `json:"keys"`
}

func (ac Account) EncodeJSON() AccountJSONMarshaler {
	return AccountJSONMarshaler{
		BaseHinter: ac.BaseHinter,
		H:          ac.h,
		AD:         ac.address,
		KS:         ac.keys,
	}
}

func (ac Account) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(AccountJSONMarshaler{
		BaseHinter: ac.BaseHinter,
		H:          ac.h,
		AD:         ac.address,
		KS:         ac.keys,
	})
}

type AccountJSONUnmarshaler struct {
	HT hint.Hint             `json:"_hint"`
	H  valuehash.HashDecoder `json:"hash"`
	AD string                `json:"address"`
	KS json.RawMessage       `json:"keys"`
}

func (ac *Account) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of Account")

	var uac AccountJSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return e(err, "")
	}

	ac.BaseHinter = hint.NewBaseHinter(uac.HT)

	switch ad, err := base.DecodeAddress(uac.AD, enc); {
	case err != nil:
		return e(err, "")
	default:
		ac.address = ad
	}

	k, err := enc.Decode(uac.KS)
	if err != nil {
		return e(err, "")
	} else if k != nil {
		v, ok := k.(BaseAccountKeys)
		if !ok {
			return util.ErrWrongType.Errorf("expected Keys, not %T", k)
		}
		ac.keys = v
	}

	ac.h = uac.H.Hash()

	return nil
}
