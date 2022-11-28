package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type CurrencyDesignJSONMarshaler struct {
	hint.BaseHinter
	AM Amount         `json:"amount"`
	GA base.Address   `json:"genesis_account"`
	PO CurrencyPolicy `json:"policy"`
	AG Big            `json:"aggregate"`
}

func (de CurrencyDesign) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(CurrencyDesignJSONMarshaler{
		BaseHinter: de.BaseHinter,
		AM:         de.amount,
		GA:         de.genesisAccount,
		PO:         de.policy,
		AG:         de.aggregate,
	})
}

type CurrencyDesignJSONUnmarshaler struct {
	AM json.RawMessage `json:"amount"`
	GA string          `json:"genesis_account"`
	PO json.RawMessage `json:"policy"`
	AG Big             `json:"aggregate"`
}

func (de *CurrencyDesign) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to unmarshal json of CurrencyDesign")

	var ude CurrencyDesignJSONUnmarshaler
	if err := enc.Unmarshal(b, &ude); err != nil {
		return e(err, "")
	}

	return de.unpack(enc, ude.AM, ude.GA, ude.PO, ude.AG)
}
