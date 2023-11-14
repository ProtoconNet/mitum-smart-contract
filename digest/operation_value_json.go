package digest

import (
	"encoding/json"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util"
	"github.com/ProtoconNet/mitum2/util/encoder"
	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/ProtoconNet/mitum2/util/localtime"
)

type OperationValueJSONMarshaler struct {
	hint.BaseHinter
	Hash        util.Hash      `json:"hash"`
	Operation   base.Operation `json:"operation"`
	Height      base.Height    `json:"height"`
	ConfirmedAt localtime.Time `json:"confirmed_at"`
	Reason      string         `json:"reason"`
	InState     bool           `json:"in_state"`
	Index       uint64         `json:"index"`
}

func (va OperationValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(OperationValueJSONMarshaler{
		BaseHinter:  va.BaseHinter,
		Hash:        va.op.Fact().Hash(),
		Operation:   va.op,
		Height:      va.height,
		ConfirmedAt: localtime.New(va.confirmedAt),
		Reason:      va.reason,
		InState:     va.inState,
		Index:       va.index,
	})
}

type OperationValueJSONUnmarshaler struct {
	Operation   json.RawMessage `json:"operation"`
	Height      base.Height     `json:"height"`
	ConfirmedAt localtime.Time  `json:"confirmed_at"`
	InState     bool            `json:"in_state"`
	Reason      string          `json:"reason"`
	Index       uint64          `json:"index"`
}

func (va *OperationValue) DecodeJSON(b []byte, enc encoder.Encoder) error {
	var uva OperationValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	if err := enc.Unmarshal(uva.Operation, &va.op); err != nil {
		return err
	}

	va.reason = uva.Reason
	va.height = uva.Height
	va.confirmedAt = uva.ConfirmedAt.Time
	va.inState = uva.InState
	va.index = uva.Index

	return nil
}
