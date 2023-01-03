package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type GenesisCurrenciesFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	GK base.Publickey   `json:"genesis_node_key"`
	KS AccountKeys      `json:"keys"`
	CS []CurrencyDesign `json:"currencies"`
}

func (fact GenesisCurrenciesFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(GenesisCurrenciesFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		GK:                    fact.genesisNodeKey,
		KS:                    fact.keys,
		CS:                    fact.cs,
	})
}

type GenesisCurrenciesFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	GK string          `json:"genesis_node_key"`
	KS json.RawMessage `json:"keys"`
	CS json.RawMessage `json:"currencies"`
}

func (fact *GenesisCurrenciesFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of GenesisCurrenciesFact")

	var uf GenesisCurrenciesFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.GK, uf.KS, uf.CS)
}

func (op GenesisCurrencies) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(op.BaseOperation)
}
