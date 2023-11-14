package extension

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum-currency/v3/types"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
)

type BaseWithdrawItemJSONMarshaler struct {
	hint.BaseHinter
	Target  base.Address   `json:"target"`
	Amounts []types.Amount `json:"amounts"`
}

func (it BaseWithdrawItem) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(BaseWithdrawItemJSONMarshaler{
		BaseHinter: it.BaseHinter,
		Target:     it.target,
		Amounts:    it.amounts,
	})
}

type BaseWithdrawItemJSONUnmarshaler struct {
	Hint    hint.Hint       `json:"_hint"`
	Target  string          `json:"target"`
	Amounts json.RawMessage `json:"amounts"`
}

func (it *BaseWithdrawItem) DecodeJSON(b []byte, enc encoder.Encoder) error {
	e := util.StringError("decode json of BaseWithdrawItem")

	var uit BaseWithdrawItemJSONUnmarshaler
	if err := enc.Unmarshal(b, &uit); err != nil {
		return e.Wrap(err)
	}

	return it.unpack(enc, uit.Hint, uit.Target, uit.Amounts)
}
