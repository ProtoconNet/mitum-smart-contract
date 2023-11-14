package types

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type CurrencyDesignJSONMarshaler struct {
	hint.BaseHinter
	Amount    Amount         `json:"amount"`
	Genesis   base.Address   `json:"genesis_account"`
	Policy    CurrencyPolicy `json:"policy"`
	Aggregate string         `json:"aggregate"`
}

func (de CurrencyDesign) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyDesignJSONMarshaler{
		BaseHinter: de.BaseHinter,
		Amount:     de.amount,
		Genesis:    de.genesisAccount,
		Policy:     de.policy,
		Aggregate:  de.aggregate.String(),
	})
}

type CurrencyDesignJSONUnmarshaler struct {
	Hint      hint.Hint       `json:"_hint"`
	Amount    json.RawMessage `json:"amount"`
	Genesis   string          `json:"genesis_account"`
	Policy    json.RawMessage `json:"policy"`
	Aggregate string          `json:"aggregate"`
}

func (de *CurrencyDesign) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of CurrencyDesign")

	var ude CurrencyDesignJSONUnmarshaler
	if err := enc.Unmarshal(b, &ude); err != nil {
		return e.Wrap(err)
	}

	return de.unpack(enc, ude.Hint, ude.Amount, ude.Genesis, ude.Policy, ude.Aggregate)
}
