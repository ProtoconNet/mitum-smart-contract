package currency

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
)

type TransfersItemJSONPacker struct {
	hint.BaseHinter
	RC base.Address `json:"receiver"`
	AM []Amount     `json:"amounts"`
}

func (it BaseTransfersItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransfersItemJSONPacker{
		BaseHinter: it.BaseHinter,
		RC:         it.receiver,
		AM:         it.amounts,
	})
}

type BaseTransfersItemJSONUnpacker struct {
	RC string          `json:"receiver"`
	AM json.RawMessage `json:"amounts"`
}

func (it *BaseTransfersItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode BaseTransfersItem")

	var utf BaseTransfersItemJSONUnpacker
	if err := enc.Unmarshal(b, &utf); err != nil {
		return e(err, "")
	}

	switch a, err := base.DecodeAddress(utf.RC, enc); {
	case err != nil:
		return e(err, "")
	default:
		it.receiver = a
	}

	ham, err := enc.DecodeSlice(utf.AM)
	if err != nil {
		return err
	}

	amounts := make([]Amount, len(ham))
	for i := range ham {
		j, ok := ham[i].(Amount)
		if !ok {
			return util.ErrWrongType.Errorf("expected Amount, not %T", ham[i])
		}

		amounts[i] = j
	}

	it.amounts = amounts

	return nil
}
