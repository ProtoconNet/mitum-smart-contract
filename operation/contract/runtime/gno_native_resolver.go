package runtime

import (
	"encoding/hex"
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/imfact-labs/mitum2/base"
	pstate "github.com/imfact-labs/smart-contract-model/state"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func CombineNativeResolvers(resolvers ...gno.NativeResolver) gno.NativeResolver {
	return func(pkgPath string, name gno.Name) func(m *gno.Machine) {
		for _, resolver := range resolvers {
			if resolver == nil {
				continue
			}

			if fn := resolver(pkgPath, name); fn != nil {
				return fn
			}
		}

		return nil
	}
}

func MitumNativeResolver(pkgPath string, name gno.Name) func(m *gno.Machine) {
	if pkgPath != MitumChainPackagePath {
		return nil
	}

	switch string(name) {
	case "AccountExists":
		return nativeAccountExists
	case "IsContractAccount":
		return nativeIsContractAccount
	case "BalanceOf":
		return nativeBalanceOf
	case "SHA3Sum256":
		return nativeSHA3Sum256
	case "CallContract":
		return nativeCallContract
	default:
		return nil
	}
}

func sha3Sum256HexString(data string) string {
	sum := sha3.Sum256([]byte(data))

	return hex.EncodeToString(sum[:])
}

func mustExecutionContext(m *gno.Machine) *ExecutionContext {
	ctx, ok := m.Context.(*ExecutionContext)
	if !ok || ctx == nil {
		panic("mitum execution context is missing from machine")
	}

	if err := ctx.Validate(); err != nil {
		panic(err)
	}

	return ctx
}

func machineStringArg(m *gno.Machine, argIndex int) string {
	b := m.LastBlock()

	var out string
	rv := reflect.ValueOf(&out).Elem()

	tv := b.GetPointerTo(nil, gno.NewValuePathBlock(1, uint16(argIndex), "")).TV
	tv.DeepFill(m.Store)
	gno.Gno2GoValue(tv, rv)

	return out
}

func machineStringMapArg(m *gno.Machine, argIndex int) map[string]string {
	b := m.LastBlock()

	tv := b.GetPointerTo(nil, gno.NewValuePathBlock(1, uint16(argIndex), "")).TV
	if tv.V == nil {
		return nil
	}
	mv, ok := tv.V.(*gno.MapValue)
	if !ok || mv.List == nil {
		panic("CallContract callData must be map[string]string")
	}

	out := make(map[string]string, mv.List.Size)
	for item := mv.List.Head; item != nil; item = item.Next {
		key := item.Key
		value := item.Value
		key.DeepFill(m.Store)
		value.DeepFill(m.Store)
		out[key.GetString()] = value.GetString()
	}

	return out
}

func pushBoolResult(m *gno.Machine, v bool) {
	m.PushValue(
		gno.Go2GnoValue(
			m.Alloc,
			m.Store,
			reflect.ValueOf(&v).Elem(),
		),
	)
}

func pushStringResult(m *gno.Machine, v string) {
	m.PushValue(
		gno.Go2GnoValue(
			m.Alloc,
			m.Store,
			reflect.ValueOf(&v).Elem(),
		),
	)
}

func pushNilErrorResult(m *gno.Machine) {
	m.PushValue(gno.TypedValue{})
}

func nativeAccountExists(m *gno.Machine) {
	ctx := mustExecutionContext(m)
	addr := machineStringArg(m, 0)

	ok, err := ctx.AccountReader.AccountExists(addr)
	if err != nil {
		panic(errors.Wrap(err, "AccountExists native call failed"))
	}

	pushBoolResult(m, ok)
}

func nativeBalanceOf(m *gno.Machine) {
	ctx := mustExecutionContext(m)
	addr := machineStringArg(m, 0)
	currency := machineStringArg(m, 1)

	value, ok, err := ctx.BalanceReader.BalanceOf(addr, currency)
	if err != nil {
		panic(errors.Wrap(err, "BalanceOf native call failed"))
	}

	pushStringResult(m, value)
	pushBoolResult(m, ok)
}

func nativeIsContractAccount(m *gno.Machine) {
	ctx := mustExecutionContext(m)
	addr := machineStringArg(m, 0)

	ok, err := ctx.ContractReader.IsContractAccount(addr)
	if err != nil {
		panic(errors.Wrap(err, "IsContractAccount native call failed"))
	}

	pushBoolResult(m, ok)
}

func nativeSHA3Sum256(m *gno.Machine) {
	data := machineStringArg(m, 0)

	pushStringResult(m, sha3Sum256HexString(data))
}

func nativeCallContract(m *gno.Machine) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(sanitizedExecutionError); ok {
				panic(r)
			}
			panic(sanitizedExecutionError{message: "chain.CallContract native failed"})
		}
	}()

	ctx := mustExecutionContext(m)
	contract := machineStringArg(m, 1)
	function := machineStringArg(m, 2)
	callData := machineStringMapArg(m, 3)

	if err := ctx.CallContract(contract, function, callData); err != nil {
		panic(sanitizedExecutionError{message: err.Error()})
	}

	pushNilErrorResult(m)
}

func (ctx *ExecutionContext) CallContract(
	contract string,
	function string,
	callData map[string]string,
) error {
	session := ctx.Session
	if session == nil {
		return errors.Errorf("chain.CallContract is unavailable outside write execution")
	}
	if ctx.ReadOnly {
		return errors.Errorf("chain.CallContract is unavailable from QueryContext")
	}
	if function == "Initialize" {
		return errors.Errorf("nested Initialize rejected for target %s at depth %d", contract, session.depth+1)
	}

	target, err := base.DecodeAddress(contract, session.encs.JSON())
	if err != nil {
		return errors.Errorf("failed to decode nested target contract %q", contract)
	}
	if ctx.Contract.Equal(target) {
		return errors.Errorf("reentrancy rejected for target %s at depth %d", target, session.depth+1)
	}

	if err := session.enterNested(target); err != nil {
		return err
	}
	defer session.leaveNested()

	designState, found, err := session.baseGetState(pstate.DesignStateKey(target))
	switch {
	case err != nil:
		return errors.Errorf("failed to read target design for contract %s function %s depth %d", target, function, session.depth)
	case !found:
		return errors.Errorf("target design not found for contract %s function %s depth %d", target, function, session.depth)
	}
	designValue, err := pstate.GetDesignStateValueFromState(designState)
	if err != nil {
		return errors.Errorf("failed to decode target design for contract %s function %s depth %d", target, function, session.depth)
	}

	runtimeState, found, err := session.getState(pstate.RuntimeStateKey(target))
	switch {
	case err != nil:
		return errors.Errorf("failed to read target runtime for contract %s function %s depth %d", target, function, session.depth)
	case !found:
		return errors.Errorf("target runtime not found for contract %s function %s depth %d", target, function, session.depth)
	}
	runtimeValue, err := pstate.GetRuntimeFromState(runtimeState)
	if err != nil {
		return errors.Errorf("failed to decode target runtime for contract %s function %s depth %d", target, function, session.depth)
	}
	if runtimeValue.Engine != pstate.RuntimeEngineGnoSnapshot {
		return errors.Errorf("target runtime mismatch for contract %s function %s depth %d", target, function, session.depth)
	}

	if _, found, err := session.snapshotFor(target); err != nil {
		return errors.Errorf("failed to read target snapshot for contract %s function %s depth %d", target, function, session.depth)
	} else if !found {
		return errors.Errorf("target snapshot not found for contract %s function %s depth %d", target, function, session.depth)
	}

	schema, err := resolveContractSchemaForExecution(nil, designValue.Design.ContractCode())
	if err != nil {
		return errors.Errorf("failed to analyze target schema for contract %s function %s depth %d", target, function, session.depth)
	}
	fn, found := schema.FindFunction(function)
	if !found || !fn.IsTypedWriteShape() {
		return errors.Errorf("target write function not found for contract %s function %s depth %d", target, function, session.depth)
	}

	err = func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				if sanitized, ok := r.(sanitizedExecutionError); ok {
					err = sanitized
					return
				}
				if session.gasMeter != nil && session.gasMeter.IsOutOfGas() {
					err = errors.Errorf("nested execution out of gas for contract %s function %s depth %d", target, function, session.depth)
					return
				}

				err = errors.Errorf("nested execution failed for contract %s function %s depth %d", target, function, session.depth)
			}
		}()

		_, err = executeContractInSession(session, ExecuteRequest{
			Mode:         InvocationModeCall,
			Contract:     target,
			Sender:       session.sender,
			Caller:       ctx.Contract,
			Height:       session.height,
			BlockTime:    session.blockTime,
			ContractCode: designValue.Design.ContractCode(),
			Schema:       &schema,
			Function:     function,
			CallData:     callData,
		}, false)

		return err
	}()
	if err != nil {
		if session.gasMeter != nil && session.gasMeter.IsOutOfGas() {
			return errors.Errorf("nested execution out of gas for contract %s function %s depth %d", target, function, session.depth)
		}
		return errors.Errorf("nested execution failed for contract %s function %s depth %d: %v", target, function, session.depth, err)
	}

	return nil
}
