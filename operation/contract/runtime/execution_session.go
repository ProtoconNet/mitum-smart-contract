package runtime

import (
	"fmt"
	"sort"

	gstore "github.com/gnolang/gno/tm2/pkg/store"
	cstate "github.com/imfact-labs/currency-model/state"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	pstate "github.com/imfact-labs/smart-contract-model/state"
)

const (
	MaxContractCallDepth            = 8
	MaxTouchedContractsPerOperation = 32
)

type contractRuntimeOverlay struct {
	snapshot []byte
}

type ExecutionSession struct {
	encs         encoder.Encoders
	baseGetState base.GetStateFunc
	gasMeter     gstore.GasMeter
	limits       GnoExecutionLimits

	topLevelMode InvocationMode
	sender       base.Address
	height       base.Height
	blockTime    int64

	depth     int
	callStack []base.Address
	touched   map[string]struct{}
	overlay   map[string]contractRuntimeOverlay
}

func NewExecutionSession(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req ExecuteRequest,
	limits GnoExecutionLimits,
	gasMeter gstore.GasMeter,
) *ExecutionSession {
	session := &ExecutionSession{
		encs:         encs,
		baseGetState: getStateFunc,
		gasMeter:     gasMeter,
		limits:       limits,
		topLevelMode: req.Mode,
		sender:       req.Sender,
		height:       req.Height,
		blockTime:    req.BlockTime,
		depth:        1,
		callStack:    []base.Address{req.Contract},
		touched:      map[string]struct{}{},
		overlay:      map[string]contractRuntimeOverlay{},
	}
	session.touched[req.Contract.String()] = struct{}{}

	return session
}

func (session *ExecutionSession) overlayGetStateFunc() base.GetStateFunc {
	return func(key string) (base.State, bool, error) {
		return session.getState(key)
	}
}

func (session *ExecutionSession) getState(key string) (base.State, bool, error) {
	if ov, found := session.overlay[key]; found {
		st, err := cstate.NewStateMergeValue(
			key,
			pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, ov.snapshot),
		).Merger(session.height, nil).CloseValue()
		return st, err == nil, err
	}

	return session.baseGetState(key)
}

func (session *ExecutionSession) snapshotFor(contract base.Address) (pstate.SnapshotStateValue, bool, error) {
	key := pstate.SnapshotStateKey(contract)
	if ov, found := session.overlay[key]; found {
		return pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, ov.snapshot), true, nil
	}

	st, found, err := session.baseGetState(key)
	if err != nil || !found {
		return pstate.SnapshotStateValue{}, found, err
	}

	snapshotValue, err := pstate.GetSnapshotFromState(st)
	return snapshotValue, true, err
}

func (session *ExecutionSession) putSnapshot(contract base.Address, snapshot []byte) {
	key := pstate.SnapshotStateKey(contract)
	session.overlay[key] = contractRuntimeOverlay{snapshot: append([]byte(nil), snapshot...)}
}

func (session *ExecutionSession) stateMerges() []base.StateMergeValue {
	keys := make([]string, 0, len(session.overlay))
	for key := range session.overlay {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	merges := make([]base.StateMergeValue, 0, len(keys))
	for _, key := range keys {
		ov := session.overlay[key]
		merges = append(merges, cstate.NewStateMergeValue(
			key,
			pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, ov.snapshot),
		))
	}

	return merges
}

func (session *ExecutionSession) enterNested(target base.Address) error {
	nextDepth := session.depth + 1
	targetString := target.String()

	if session.topLevelMode == InvocationModeRegister {
		return fmt.Errorf("nested call during RegisterContract rejected for target %s", targetString)
	}
	if nextDepth > MaxContractCallDepth {
		return fmt.Errorf("depth exceeded for target %s at depth %d", targetString, nextDepth)
	}
	for _, stacked := range session.callStack {
		if stacked.Equal(target) {
			if stacked.Equal(session.callStack[0]) {
				return fmt.Errorf("reentrancy rejected for target %s at depth %d", targetString, nextDepth)
			}
			return fmt.Errorf("reentrancy rejected for target %s at depth %d", targetString, nextDepth)
		}
	}
	if _, found := session.touched[targetString]; !found {
		if len(session.touched)+1 > MaxTouchedContractsPerOperation {
			return fmt.Errorf("touched limit exceeded for target %s at depth %d", targetString, nextDepth)
		}
		session.touched[targetString] = struct{}{}
	}

	session.depth = nextDepth
	session.callStack = append(session.callStack, target)

	return nil
}

func (session *ExecutionSession) leaveNested() {
	if session.depth > 1 {
		session.depth--
	}
	if len(session.callStack) > 1 {
		session.callStack = session.callStack[:len(session.callStack)-1]
	}
}
