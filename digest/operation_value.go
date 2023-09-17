package digest

import (
	"time"

	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/hint"
)

var (
	OperationValueHint = hint.MustNewHint("mitum-currency-operation-value-v0.0.1")
)

type OperationValue struct {
	hint.BaseHinter
	op          base.Operation
	height      base.Height
	confirmedAt time.Time
	inState     bool
	reason      string
	index       uint64
}

func NewOperationValue(
	op base.Operation,
	height base.Height,
	confirmedAt time.Time,
	inState bool,
	reason string,
	index uint64,
) OperationValue {
	return OperationValue{
		BaseHinter:  hint.NewBaseHinter(OperationValueHint),
		op:          op,
		height:      height,
		confirmedAt: confirmedAt,
		inState:     inState,
		reason:      reason,
		index:       index,
	}
}

func (OperationValue) Hint() hint.Hint {
	return OperationValueHint
}

func (va OperationValue) Operation() base.Operation {
	return va.op
}

func (va OperationValue) Height() base.Height {
	return va.height
}

func (va OperationValue) ConfirmedAt() time.Time {
	return va.confirmedAt
}

func (va OperationValue) InState() bool {
	return va.inState
}

func (va OperationValue) Reason() string {
	return va.reason
}

// Index indicates the index number of Operation in OperationTree of block.
func (va OperationValue) Index() uint64 {
	return va.index
}
