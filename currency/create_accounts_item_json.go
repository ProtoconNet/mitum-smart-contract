package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type CreateAccountsItemJSONMarshaler struct {
	hint.BaseHinter
	KS AccountKeys `json:"keys"`
	AS []Amount    `json:"amounts"`
}

func (it BaseCreateAccountsItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateAccountsItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		KS:         it.keys,
		AS:         it.amounts,
	})
}

type CreateAccountsItemJSONUnMarshaler struct {
	HT hint.Hint       `json:"_hint"`
	KS json.RawMessage `json:"keys"`
	AM json.RawMessage `json:"amounts"`
}

func (it *BaseCreateAccountsItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of BaseCreateAccountsItem")

	var uit CreateAccountsItemJSONUnMarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.HT, uit.KS, uit.AM)
}
