package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type TransfersItemJSONPacker struct {
	hint.BaseHinter
	Receiver base.Address   `json:"receiver"`
	Amounts  []types.Amount `json:"amounts"`
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
	e := util.StringError("failed to decode json of BaseTransfersItem")

	var uit BaseTransfersItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, uit.Hint, uit.Receiver, uit.Amounts)
}
