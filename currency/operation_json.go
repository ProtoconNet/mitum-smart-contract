package currency

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
)

func (op BaseOperation) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"_hint": op.Hint(),
		"hash":  op.Hash(),
		"fact":  op.Fact(),
		"signs": op.Signs(),
		"memo":  op.Memo,
	}

	return util.MarshalJSON(m)
}

func (op *BaseOperation) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var ubo base.BaseOperation
	err := ubo.DecodeJSON(b, enc)
	if err != nil {
		return err
	}

	op.BaseOperation = ubo

	var um MemoJSONUnMarshaler
	if err := enc.Unmarshal(b, &um); err != nil {
		return err
	}
	op.Memo = um.Memo

	return nil
}
