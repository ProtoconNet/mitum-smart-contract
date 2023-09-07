package currency

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
)

type RegisterGenesisCurrencyFactJSONMarshaler struct {
	base.BaseFactJSONMarshaler
	GenesisNodeKey base.Publickey         `json:"genesis_node_key"`
	Keys           types.AccountKeys      `json:"keys"`
	Currencies     []types.CurrencyDesign `json:"currencies"`
}

func (fact RegisterGenesisCurrencyFact) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(RegisterGenesisCurrencyFactJSONMarshaler{
		BaseFactJSONMarshaler: fact.BaseFact.JSONMarshaler(),
		GenesisNodeKey:        fact.genesisNodeKey,
		Keys:                  fact.keys,
		Currencies:            fact.cs,
	})
}

type RegisterGenesisCurrencyFactJSONUnMarshaler struct {
	base.BaseFactJSONUnmarshaler
	GenesisNodeKey string          `json:"genesis_node_key"`
	Keys           json.RawMessage `json:"keys"`
	Currencies     json.RawMessage `json:"currencies"`
}

func (fact *RegisterGenesisCurrencyFact) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("decode json of RegisterGenesisCurrencyFact")

	var uf RegisterGenesisCurrencyFactJSONUnMarshaler
	if err := enc.Unmarshal(b, &uf); err != nil {
		return e.Wrap(err)
	}

	fact.BaseFact.SetJSONUnmarshaler(uf.BaseFactJSONUnmarshaler)

	return fact.unpack(enc, uf.GenesisNodeKey, uf.Keys, uf.Currencies)
}

func (op RegisterGenesisCurrency) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(op.BaseOperation)
}
