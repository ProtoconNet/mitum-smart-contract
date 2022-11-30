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

	var uca CreateAccountsItemJSONUnMarshaler
	if err := enc.Unmarshal(b, &uca); err != nil {
		return e(err, "")
	}

	it.BaseHinter = hint.NewBaseHinter(uca.HT)

	if hinter, err := enc.Decode(uca.KS); err != nil {
		return err
	} else if k, ok := hinter.(AccountKeys); !ok {
		return e(util.ErrWrongType.Errorf("expected AccountsKeys not %T,", hinter), "")
	} else {
		it.keys = k
	}

	ham, err := enc.DecodeSlice(uca.AM)
	if err != nil {
		return err
	}

	amounts := make([]Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(Amount)
		if !ok {
			return e(util.ErrWrongType.Errorf("expected Amount, not %T", ham[i]), "")
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
