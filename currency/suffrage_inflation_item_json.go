package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"

	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

type SuffrageInflationItemJSONMarshaler struct {
	RC base.Address `bson:"receiver" json:"receiver"`
	AM Amount       `bson:"amount" json:"amount"`
}

func (it SuffrageInflationItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(SuffrageInflationItemJSONMarshaler{
		RC: it.receiver,
		AM: it.amount,
	})
}

type SuffrageInflationItemJSONUnmarshaler struct {
	RC string          `bson:"receiver" json:"receiver"`
	AM json.RawMessage `bson:"amount" json:"amount"`
}

func (it *SuffrageInflationItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of SuffrageInflationItem")

	var uit SuffrageInflationItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return err
	}

	switch a, err := base.DecodeAddress(uit.RC, enc); {
	case err != nil:
		return err
	default:
		it.receiver = a
	}

	if hinter, err := enc.Decode(uit.AM); err != nil {
		return err
	} else if am, ok := hinter.(Amount); !ok {
		return e(util.ErrWrongType.Errorf("expected Amount not %T,", hinter), "")
	} else {
		it.amount = am
	}

	return nil
}
