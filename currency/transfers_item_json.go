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
	Receiver base.Address `json:"receiver"`
	Amounts  []Amount     `json:"amounts"`
}

func (it BaseTransfersItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(TransfersItemJSONPacker{
		BaseHinter: it.BaseHinter,
		Receiver:   it.receiver,
		Amounts:    it.amounts,
	})
}

type BaseTransfersItemJSONUnpacker struct {
	Hint     hint.Hint       `json:"_hint"`
	Receiver string          `json:"receiver"`
	Amounts  json.RawMessage `json:"amounts"`
}

func (it *BaseTransfersItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringErrorFunc("failed to decode json of BaseTransfersItem")

	var uit BaseTransfersItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e(err, "")
	}

	return it.unpack(enc, uit.Hint, uit.Receiver, uit.Amounts)
}
