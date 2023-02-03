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
	e := util.StringErrorFunc("failed to decode json of Account")

	var uac AccountJSONUnmarshaler
	if err := enc.Unmarshal(b, &uac); err != nil {
		return e(err, "")
	}

	return ac.unpack(enc, uac.HT, uac.H, uac.AD, uac.KS)
}
