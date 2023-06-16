package extension

import (
	"encoding/json"
	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	jsonenc "github.com/ProtoconNet/mitum2/util/encoder/json"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type WithdrawsItemJSONMarshaler struct {
	hint.BaseHinter
	Target  base.Address   `json:"target"`
	Amounts []types.Amount `json:"amounts"`
}

func (it BaseWithdrawsItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(WithdrawsItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Target:     it.target,
		Amounts:    it.amounts,
	})
}

type BaseWithdrawsItemJSONUnpacker struct {
	Hint    hint.Hint       `json:"_hint"`
	Target  string          `json:"target"`
	Amounts json.RawMessage `json:"amounts"`
}

func (it *BaseWithdrawsItem) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	e := util.StringError("failed to decode json of BaseWithdrawsItem")

	var uit BaseWithdrawsItemJSONUnpacker
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, uit.Hint, uit.Target, uit.Amounts)
}
