package runtime

import (
	"fmt"
	"sort"

	gstore "github.com/gnolang/gno/tm2/pkg/store"
	ccommon "github.com/imfact-labs/currency-model/common"
	cstate "github.com/imfact-labs/currency-model/state"
	ccurrency "github.com/imfact-labs/currency-model/state/currency"
	ctypes "github.com/imfact-labs/currency-model/types"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util"
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

type currencyBalanceOverlay struct {
	address  base.Address
	currency ctypes.CurrencyID
	add      ccommon.Big
	remove   ccommon.Big
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

	accountOverlay map[string]base.StateMergeValue
	balanceOverlay map[string]currencyBalanceOverlay
}

func NewExecutionSession(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req ExecuteRequest,
	limits GnoExecutionLimits,
	gasMeter gstore.GasMeter,
) *ExecutionSession {
	session := &ExecutionSession{
		encs:           encs,
		baseGetState:   getStateFunc,
		gasMeter:       gasMeter,
		limits:         limits,
		topLevelMode:   req.Mode,
		sender:         req.Sender,
		height:         req.Height,
		blockTime:      req.BlockTime,
		depth:          1,
		callStack:      []base.Address{req.Contract},
		touched:        map[string]struct{}{},
		overlay:        map[string]contractRuntimeOverlay{},
		accountOverlay: map[string]base.StateMergeValue{},
		balanceOverlay: map[string]currencyBalanceOverlay{},
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
	if smv, found := session.accountOverlay[key]; found {
		merger := smv.Merger(session.height, nil)
		if err := merger.Merge(smv.Value(), nil); err != nil {
			return nil, false, err
		}
		st, err := merger.CloseValue()
		return st, err == nil, err
	}
	if ov, found := session.balanceOverlay[key]; found {
		balance, ok, err := session.projectedBalanceByKey(key, ov)
		if err != nil || !ok {
			return nil, false, err
		}
		return ccommon.NewBaseState(
			session.height,
			key,
			ccurrency.NewBalanceStateValue(ctypes.NewAmount(balance, ov.currency)),
			nil,
			[]util.Hash{},
		), true, nil
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

	accountKeys := make([]string, 0, len(session.accountOverlay))
	for key := range session.accountOverlay {
		accountKeys = append(accountKeys, key)
	}
	sort.Strings(accountKeys)
	for _, key := range accountKeys {
		merges = append(merges, session.accountOverlay[key])
	}

	balanceKeys := make([]string, 0, len(session.balanceOverlay))
	for key := range session.balanceOverlay {
		balanceKeys = append(balanceKeys, key)
	}
	sort.Strings(balanceKeys)
	for _, key := range balanceKeys {
		ov := session.balanceOverlay[key]
		if ov.remove.OverZero() {
			merges = append(merges, newBalanceStateMergeValue(
				key,
				ov.currency,
				ccurrency.NewDeductBalanceStateValue(ctypes.NewAmount(ov.remove, ov.currency)),
			))
		}
		if ov.add.OverZero() {
			merges = append(merges, newBalanceStateMergeValue(
				key,
				ov.currency,
				ccurrency.NewAddBalanceStateValue(ctypes.NewAmount(ov.add, ov.currency)),
			))
		}
	}

	return merges
}

func (session *ExecutionSession) putAccountMerge(smv base.StateMergeValue) {
	session.accountOverlay[smv.Key()] = smv
}

func (session *ExecutionSession) accountExists(address base.Address) (bool, error) {
	key := ccurrency.AccountStateKey(address)
	_, found, err := session.getState(key)

	return found, err
}

func (session *ExecutionSession) addBalance(address base.Address, cid ctypes.CurrencyID, amount ccommon.Big) {
	key := ccurrency.BalanceStateKey(address, cid)
	ov := session.balanceOverlay[key]
	if ov.currency == "" {
		ov = currencyBalanceOverlay{
			address:  address,
			currency: cid,
			add:      ccommon.ZeroBig,
			remove:   ccommon.ZeroBig,
		}
	}
	ov.add = ov.add.Add(amount)
	session.balanceOverlay[key] = ov
}

func (session *ExecutionSession) removeBalance(address base.Address, cid ctypes.CurrencyID, amount ccommon.Big) {
	key := ccurrency.BalanceStateKey(address, cid)
	ov := session.balanceOverlay[key]
	if ov.currency == "" {
		ov = currencyBalanceOverlay{
			address:  address,
			currency: cid,
			add:      ccommon.ZeroBig,
			remove:   ccommon.ZeroBig,
		}
	}
	ov.remove = ov.remove.Add(amount)
	session.balanceOverlay[key] = ov
}

func (session *ExecutionSession) projectedBalance(address base.Address, cid ctypes.CurrencyID) (ccommon.Big, bool, error) {
	key := ccurrency.BalanceStateKey(address, cid)
	if ov, found := session.balanceOverlay[key]; found {
		return session.projectedBalanceByKey(key, ov)
	}

	st, found, err := session.baseGetState(key)
	if err != nil || !found {
		return ccommon.ZeroBig, found, err
	}

	amount, err := ccurrency.StateBalanceValue(st)
	if err != nil {
		return ccommon.ZeroBig, false, err
	}

	return amount.Big(), true, nil
}

func (session *ExecutionSession) projectedBalanceByKey(key string, ov currencyBalanceOverlay) (ccommon.Big, bool, error) {
	balance := ccommon.ZeroBig
	if st, found, err := session.baseGetState(key); err != nil {
		return ccommon.ZeroBig, false, err
	} else if found {
		amount, err := ccurrency.StateBalanceValue(st)
		if err != nil {
			return ccommon.ZeroBig, false, err
		}
		balance = amount.Big()
	}

	balance = balance.Add(ov.add).Sub(ov.remove)
	if !balance.OverNil() {
		return ccommon.ZeroBig, false, fmt.Errorf("projected balance underflow for %s", key)
	}

	return balance, true, nil
}

func newBalanceStateMergeValue(
	key string,
	cid ctypes.CurrencyID,
	value base.StateValue,
) base.StateMergeValue {
	return ccommon.NewBaseStateMergeValue(
		key,
		value,
		func(height base.Height, st base.State) base.StateValueMerger {
			return ccurrency.NewBalanceStateValueMerger(height, key, cid, st)
		},
	)
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
