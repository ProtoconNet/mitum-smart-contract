package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type CreateAccountsFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	SD base.Address         `json:"sender"`
	IT []CreateAccountsItem `json:"items"`
}

func (fact CreateAccountsFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CreateAccountsFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		SD:                    fact.sender,
		IT:                    fact.items,
	})
}

type CreateAccountsFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	SD string          `json:"sender"`
	IT json.RawMessage `json:"items"`
}

func (fact *CreateAccountsFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode CreateAccountsFact")

	var uca CreateAccountsFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uca); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uca.BaseFactJSONUnmarshaler)
	switch a, err := base.DecodeAddress(uca.SD, enc); {
	case err != nil:
		return e(err, "")
	default:
		fact.sender = a
	}

	hit, err := enc.DecodeSlice(uca.IT)
	if err != nil {
		return e(err, "")
	}

	items := make([]CreateAccountsItem, len(hit))
	for i := range hit {
		j, ok := hit[i].(CreateAccountsItem)
		if !ok {
			return util.ErrWrongType.Errorf("expected CreateAccountsItem, not %T", hit[i])
		}

		items[i] = j
	}
	fact.items = items

	return nil
}

func (op *CreateAccounts) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo BaseOperation
	if err := ubo.DecodeJSON(b, enc); err != nil {
		return err
	}

	op.BaseOperation = ubo

	return nil
}
