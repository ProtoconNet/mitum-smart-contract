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
	HT hint.Hint       `json:"_hint"`
	RC string          `json:"receiver"`
	AM json.RawMessage `json:"amounts"`
}

func (it *BaseTransfersItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of BaseTransfersItem")

	var uit BaseTransfersItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.HT, uit.RC, uit.AM)
}
