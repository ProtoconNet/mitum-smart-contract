package currency

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/common"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type MintItemJSONMarshaler struct {
	hint.BaseHinter
	Receiver base.Address `json:"receiver"`
	Amount   types.Amount `json:"amount"`
}

func (it MintItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(MintItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Receiver:   it.receiver,
		Amount:     it.amount,
	})
}

type MintItemJSONUnmarshaler struct {
	HT       hint.Hint       `json:"_hint"`
	Receiver string          `json:"receiver"`
	Amount   json.RawMessage `json:"amount"`
}

func (it *MintItem) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uit MintItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	if err := it.unpack(enc, uit.HT, uit.Receiver, uit.Amount); err != nil {
		return common.DecorateError(err, common.ErrDecodeJson, *it)
	}

	return nil
}
