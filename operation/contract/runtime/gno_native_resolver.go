package runtime

import (
	"encoding/hex"
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	ccommon "github.com/imfact-labs/currency-model/common"
	cstate "github.com/imfact-labs/currency-model/state"
	ccurrency "github.com/imfact-labs/currency-model/state/currency"
	cestate "github.com/imfact-labs/currency-model/state/extension"
	ctypes "github.com/imfact-labs/currency-model/types"
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
	case "TransferSenderToContract":
		return nativeTransferSenderToContract
	case "TransferContractTo":
		return nativeTransferContractTo
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

func nativeTransferSenderToContract(m *gno.Machine) {
	defer sanitizeTransferNativePanic("chain.TransferSenderToContract native failed")

	ctx := mustExecutionContext(m)
	currency := machineStringArg(m, 1)
	amount := machineStringArg(m, 2)

	if err := ctx.TransferSenderToContract(currency, amount); err != nil {
		panic(sanitizedExecutionError{message: err.Error()})
	}

	pushNilErrorResult(m)
}

func nativeTransferContractTo(m *gno.Machine) {
	defer sanitizeTransferNativePanic("chain.TransferContractTo native failed")

	ctx := mustExecutionContext(m)
	receiver := machineStringArg(m, 1)
	currency := machineStringArg(m, 2)
	amount := machineStringArg(m, 3)

	if err := ctx.TransferContractTo(receiver, currency, amount); err != nil {
		panic(sanitizedExecutionError{message: err.Error()})
	}

	pushNilErrorResult(m)
}

func sanitizeTransferNativePanic(message string) func() {
	return func() {
		if r := recover(); r != nil {
			if _, ok := r.(sanitizedExecutionError); ok {
				panic(r)
			}
			panic(sanitizedExecutionError{message: message})
		}
	}
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

func (ctx *ExecutionContext) TransferSenderToContract(currencyID string, amountText string) error {
	session, err := ctx.ensureTopLevelCurrencyTransferAllowed("TransferSenderToContract")
	if err != nil {
		return err
	}

	cid, amount, err := parseCurrencyAmount(currencyID, amountText)
	if err != nil {
		return err
	}
	if err := session.ensureCurrencyDesignExists(cid); err != nil {
		return err
	}
	if err := session.ensureRegularAccountExists(session.sender, "sender"); err != nil {
		return err
	}
	if err := session.ensureAddressIsNotContractAccount(session.sender, "sender"); err != nil {
		return err
	}
	if err := session.ensureContractAccountExists(ctx.Contract, "contract"); err != nil {
		return err
	}

	sourceBalance, found, err := session.projectedBalance(session.sender, cid)
	switch {
	case err != nil:
		return errors.Errorf("failed to read sender balance for currency %s", cid)
	case !found:
		return errors.Errorf("sender balance not found for currency %s", cid)
	case sourceBalance.Compare(amount) < 0:
		return errors.Errorf("insufficient sender balance for currency %s", cid)
	}

	session.removeBalance(session.sender, cid, amount)
	session.addBalance(ctx.Contract, cid, amount)

	return nil
}

func (ctx *ExecutionContext) TransferContractTo(receiverText string, currencyID string, amountText string) error {
	session, err := ctx.ensureTopLevelCurrencyTransferAllowed("TransferContractTo")
	if err != nil {
		return err
	}

	receiver, err := decodeRuntimeAddress(receiverText, session.encs)
	if err != nil {
		return errors.Errorf("failed to decode transfer receiver %q", receiverText)
	}
	cid, amount, err := parseCurrencyAmount(currencyID, amountText)
	if err != nil {
		return err
	}
	if err := session.ensureCurrencyDesignExists(cid); err != nil {
		return err
	}
	if err := session.ensureContractAccountExists(ctx.Contract, "contract"); err != nil {
		return err
	}

	sourceBalance, found, err := session.projectedBalance(ctx.Contract, cid)
	switch {
	case err != nil:
		return errors.Errorf("failed to read contract balance for currency %s", cid)
	case !found:
		return errors.Errorf("contract balance not found for currency %s", cid)
	case sourceBalance.Compare(amount) < 0:
		return errors.Errorf("insufficient contract balance for currency %s", cid)
	}

	if found, err := session.accountExists(receiver); err != nil {
		return errors.Errorf("failed to read receiver account %s", receiver)
	} else if !found {
		smv, err := cstate.CreateNotExistAccount(receiver, session.overlayGetStateFunc())
		if err != nil {
			return errors.Errorf("failed to create receiver account %s", receiver)
		}
		if smv != nil {
			session.putAccountMerge(smv)
		}
	}

	session.removeBalance(ctx.Contract, cid, amount)
	session.addBalance(receiver, cid, amount)

	return nil
}

func (ctx *ExecutionContext) ensureTopLevelCurrencyTransferAllowed(name string) (*ExecutionSession, error) {
	session := ctx.Session
	if session == nil {
		return nil, errors.Errorf("chain.%s is unavailable outside write execution", name)
	}
	if ctx.ReadOnly {
		return nil, errors.Errorf("chain.%s is unavailable from QueryContext", name)
	}
	if session.topLevelMode == InvocationModeRegister {
		return nil, errors.Errorf("chain.%s is not allowed during contract registration", name)
	}
	if session.topLevelMode != InvocationModeCall {
		return nil, errors.Errorf("chain.%s is allowed only during contract call execution", name)
	}
	if session.depth != 1 || len(session.callStack) < 1 || !ctx.Contract.Equal(session.callStack[0]) {
		return nil, errors.Errorf("chain.%s is allowed only from the top-level contract", name)
	}

	return session, nil
}

func parseCurrencyAmount(currencyID string, amountText string) (ctypes.CurrencyID, ccommon.Big, error) {
	cid := ctypes.CurrencyID(currencyID)
	if err := cid.IsValid(nil); err != nil {
		return "", ccommon.ZeroBig, errors.Errorf("invalid currency %q", currencyID)
	}

	amount, err := ccommon.NewBigFromString(amountText)
	if err != nil {
		return "", ccommon.ZeroBig, errors.Errorf("invalid transfer amount %q", amountText)
	}
	if !amount.OverZero() {
		return "", ccommon.ZeroBig, errors.Errorf("transfer amount must be positive")
	}

	return cid, amount, nil
}

func (session *ExecutionSession) ensureCurrencyDesignExists(cid ctypes.CurrencyID) error {
	st, found, err := session.baseGetState(ccurrency.DesignStateKey(cid))
	switch {
	case err != nil:
		return errors.Errorf("failed to read currency design %s", cid)
	case !found:
		return errors.Errorf("currency design not found for %s", cid)
	}
	if _, ok := st.Value().(ccurrency.DesignStateValue); !ok {
		return errors.Errorf("invalid currency design state for %s", cid)
	}

	return nil
}

func (session *ExecutionSession) ensureRegularAccountExists(address base.Address, name string) error {
	st, found, err := session.getState(ccurrency.AccountStateKey(address))
	switch {
	case err != nil:
		return errors.Errorf("failed to read %s account %s", name, address)
	case !found:
		return errors.Errorf("%s account not found %s", name, address)
	}
	if _, err := ccurrency.LoadAccountStateValue(st); err != nil {
		return errors.Errorf("invalid %s account state %s", name, address)
	}

	return nil
}

func (session *ExecutionSession) ensureContractAccountExists(address base.Address, name string) error {
	st, found, err := session.baseGetState(cestate.StateKeyContractAccount(address))
	switch {
	case err != nil:
		return errors.Errorf("failed to read %s contract account %s", name, address)
	case !found:
		return errors.Errorf("%s contract account not found %s", name, address)
	}
	if _, err := cestate.StateContractAccountValue(st); err != nil {
		return errors.Errorf("invalid %s contract account state %s", name, address)
	}

	return nil
}

func (session *ExecutionSession) ensureAddressIsNotContractAccount(address base.Address, name string) error {
	st, found, err := session.baseGetState(cestate.StateKeyContractAccount(address))
	switch {
	case err != nil:
		return errors.Errorf("failed to read %s contract account %s", name, address)
	case !found:
		return nil
	}
	if _, err := cestate.StateContractAccountValue(st); err != nil {
		return errors.Errorf("invalid %s contract account state %s", name, address)
	}

	return errors.Errorf("%s must not be a contract account %s", name, address)
}
