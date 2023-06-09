package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type GenesisCurrenciesFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	GenesisNodeKey base.Publickey         `json:"genesis_node_key"`
	Keys           types.AccountKeys      `json:"keys"`
	Currencies     []types.CurrencyDesign `json:"currencies"`
}

func (fact GenesisCurrenciesFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(GenesisCurrenciesFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		GenesisNodeKey:        fact.genesisNodeKey,
		Keys:                  fact.keys,
		Currencies:            fact.cs,
	})
}

type GenesisCurrenciesFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	GenesisNodeKey string          `json:"genesis_node_key"`
	Keys           json.RawMessage `json:"keys"`
	Currencies     json.RawMessage `json:"currencies"`
}

func (fact *GenesisCurrenciesFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of GenesisCurrenciesFact")

	var uf GenesisCurrenciesFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e(err, "")
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.GenesisNodeKey, uf.Keys, uf.Currencies)
}

func (op GenesisCurrencies) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(op.BaseOperation)
}
