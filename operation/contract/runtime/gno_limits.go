package runtime

import (
	"fmt"
	"strings"

	"github.com/ProtoconNet/mitum2/base"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
)

const (
	GnoWriteGasLimit      = int64(5_000_000)
	GnoQueryGasLimit      = int64(1_000_000)
	GnoWriteMaxAllocBytes = int64(8 << 20)
	GnoQueryMaxAllocBytes = int64(4 << 20)
)

type GnoExecutionLimits struct {
	GasLimit      int64
	MaxAllocBytes int64
}

func WriteGnoExecutionLimits() GnoExecutionLimits {
	return GnoExecutionLimits{
		GasLimit:      GnoWriteGasLimit,
		MaxAllocBytes: GnoWriteMaxAllocBytes,
	}
}

func QueryGnoExecutionLimits() GnoExecutionLimits {
	return GnoExecutionLimits{
		GasLimit:      GnoQueryGasLimit,
		MaxAllocBytes: GnoQueryMaxAllocBytes,
	}
}

func NewGnoGasMeter(limit int64) gstore.GasMeter {
	return gstore.NewGasMeter(limit)
}

func ClassifyGnoExecutionPanic(
	scope string,
	r any,
	gasMeter gstore.GasMeter,
) base.OperationProcessReasonError {
	if gasMeter != nil && gasMeter.IsOutOfGas() {
		return base.NewBaseOperationProcessReasonError(
			"%s out of gas", scope,
		)
	}

	msg := fmt.Sprint(r)
	lmsg := strings.ToLower(msg)

	if strings.Contains(lmsg, "alloc") || strings.Contains(lmsg, "allocation") {
		return base.NewBaseOperationProcessReasonError(
			"%s exceeded allocation limit: %v", scope, r,
		)
	}

	return base.NewBaseOperationProcessReasonError(
		"%s panicked: %v", scope, r,
	)
}
