package digest

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/util"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/localtime"
)

type OperationValueJSONMarshaler struct {
	hint.BaseHinter
	HS util.Hash                        `json:"hash"`
	OP base.Operation                   `json:"operation"`
	HT base.Height                      `json:"height"`
	CF localtime.Time                   `json:"confirmed_at"`
	RS base.OperationProcessReasonError `json:"reason"`
	IN bool                             `json:"in_state"`
	ID uint64                           `json:"index"`
}

func (va OperationValue) MarshalJSON() ([]byte, error) {
	return util.MarshalJSON(OperationValueJSONMarshaler{
		BaseHinter: va.BaseHinter,
		HS:         va.op.Fact().Hash(),
		OP:         va.op,
		HT:         va.height,
		CF:         localtime.New(va.confirmedAt),
		RS:         va.reason,
		IN:         va.inState,
		ID:         va.index,
	})
}

type OperationValueJSONUnmarshaler struct {
	OP json.RawMessage `json:"operation"`
	HT base.Height     `json:"height"`
	CF localtime.Time  `json:"confirmed_at"`
	IN bool            `json:"in_state"`
	RS json.RawMessage `json:"reason"`
	ID uint64          `json:"index"`
}

func (va *OperationValue) DecodeJSON(b []byte, enc *jsonenc.Encoder) error {
	var uva OperationValueJSONUnmarshaler
	if err := enc.Unmarshal(b, &uva); err != nil {
		return err
	}

	if err := enc.Unmarshal(uva.OP, &va.op); err != nil {
		return err
	}

	if err := enc.Unmarshal(uva.RS, &va.reason); err != nil {
		return err
	}

	va.height = uva.HT
	va.confirmedAt = uva.CF.Time
	va.inState = uva.IN
	va.index = uva.ID

	return nil
}
