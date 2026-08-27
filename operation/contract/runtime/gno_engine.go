package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnostdlibs "github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	cstate "github.com/imfact-labs/currency-model/state"
	"github.com/imfact-labs/mitum2/base"
	"github.com/imfact-labs/mitum2/util/encoder"
	"github.com/imfact-labs/smart-contract-model/state"
)

type gnoEngine struct{}

var analyzeContractSchemaFunc = AnalyzeContractSchema
var newGnoMachineAndPackageFunc func(
	*ExecutionContext,
	string,
	string,
	GnoExecutionLimits,
	gstore.GasMeter,
) (*gno.Machine, *gno.PackageValue, error)

func init() {
	newGnoMachineAndPackageFunc = newGnoMachineAndPackage
}

func NewGnoEngine() ContractEngine {
	return gnoEngine{}
}

func (gnoEngine) ValidateContract(sourceCode string) (ContractSchema, base.OperationProcessReasonError) {
	schema, err := resolveContractSchemaForExecution(nil, sourceCode)
	if err != nil {
		return ContractSchema{}, base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	return schema, nil
}

func (gnoEngine) ExecuteContract(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req ExecuteRequest,
) (_ ExecuteResult, berr base.OperationProcessReasonError) {
	var gasMeter gstore.GasMeter

	defer func() {
		if r := recover(); r != nil {
			berr = ClassifyGnoExecutionPanic("gno execution", r, gasMeter)
		}
	}()

	if _, err := normalizeExecuteRequest(req); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("invalid call data: %v", err)
	}

	limits := WriteGnoExecutionLimits()
	gasMeter = NewGnoGasMeter(limits.GasLimit)
	session := NewExecutionSession(encs, getStateFunc, req, limits, gasMeter)

	result, err := executeContractInSession(session, req, true)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("%v", err)
	}

	return result, nil
}

func executeContractInSession(
	session *ExecutionSession,
	req ExecuteRequest,
	topLevel bool,
) (ExecuteResult, error) {
	normalized, err := normalizeExecuteRequest(req)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("invalid call data: %v", err)
	}

	schema, err := resolveContractSchemaForExecution(req.Schema, req.ContractCode)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to analyze contract schema: %v", err)
	}

	var runtimeValue state.RuntimeStateValue
	isNewRuntime := false

	getStateFunc := session.overlayGetStateFunc()

	if st, found, err := getStateFunc(state.RuntimeStateKey(req.Contract)); err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to read runtime state: %v", err)
	} else if found {
		runtimeValue, err = state.GetRuntimeFromState(st)
		if err != nil {
			return ExecuteResult{}, fmt.Errorf("failed to decode runtime state: %v", err)
		}
		if runtimeValue.Engine != state.RuntimeEngineGnoSnapshot {
			return ExecuteResult{}, fmt.Errorf(
				"runtime engine mismatch: %q", runtimeValue.Engine,
			)
		}
	} else {
		if req.Mode != InvocationModeRegister {
			return ExecuteResult{}, fmt.Errorf(
				"runtime state not found for typed contract %v; register must create runtime first",
				req.Contract,
			)
		}

		runtimeValue = deriveRuntimeState(req.Contract, req.ContractCode)
		isNewRuntime = true
	}

	var snapshotValue state.SnapshotStateValue

	if snapshotFromState, found, err := session.snapshotFor(req.Contract); err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to read snapshot state: %v", err)
	} else if found {
		snapshotValue = snapshotFromState
	} else {
		if req.Mode != InvocationModeRegister {
			return ExecuteResult{}, fmt.Errorf(
				"snapshot state not found for typed contract %v",
				req.Contract,
			)
		}

		snapshotValue = state.NewSnapshotStateValue(
			GnoSnapshotVersion,
			GnoSnapshotCodecName,
			nil,
		)
	}

	if snapshotValue.Version != GnoSnapshotVersion {
		return ExecuteResult{}, fmt.Errorf(
			"unsupported snapshot version %d", snapshotValue.Version,
		)
	}
	if snapshotValue.Codec != GnoSnapshotCodecName {
		return ExecuteResult{}, fmt.Errorf(
			"unsupported snapshot codec %q", snapshotValue.Codec,
		)
	}

	caller := req.Caller
	if caller == nil {
		caller = req.Sender
	}

	execCtx, err := NewExecutionContextWithCaller(
		session.encs,
		getStateFunc,
		req.Contract,
		req.Sender,
		caller,
		req.Height,
		false,
	)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to build execution context: %v", err)
	}
	execCtx.Session = session

	m, pkg, err := newGnoMachineAndPackageFunc(
		execCtx,
		runtimeValue.PackagePath,
		req.ContractCode,
		session.limits,
		session.gasMeter,
	)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to initialize gno machine: %v", err)
	}

	if err := RestoreSnapshot(m, pkg, snapshotValue.Snapshot, schema); err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to restore snapshot: %v", err)
	}

	switch normalized.mode {
	case InvocationModeRegister:
		invokeReq := req
		invokeReq.Mode = InvocationModeRegister
		invokeReq.Function = normalized.registerFunction
		invokeReq.CallData = normalized.initData
		if err := invokeTypedWrite(m, pkg, invokeReq, schema); err != nil {
			return ExecuteResult{}, fmt.Errorf("failed to execute typed contract call: %v", err)
		}
	case InvocationModeCall:
		for i := range normalized.callItems {
			item := normalized.callItems[i]
			if item.Function == "Initialize" {
				return ExecuteResult{}, fmt.Errorf(
					"failed to execute typed contract call: call item %d %q: Initialize cannot be called through call operation for typed contracts",
					i+1,
					item.Function,
				)
			}

			invokeReq := req
			invokeReq.Mode = InvocationModeCall
			invokeReq.Function = item.Function
			invokeReq.CallData = item.CallData
			if err := invokeTypedWrite(m, pkg, invokeReq, schema); err != nil {
				return ExecuteResult{}, fmt.Errorf(
					"failed to execute typed contract call: call item %d %q: %v",
					i+1,
					item.Function,
					err,
				)
			}
		}
	default:
		return ExecuteResult{}, fmt.Errorf("unsupported invocation mode %q", normalized.mode)
	}

	snapshotBytes, err := CaptureSnapshot(pkg, m.Store, schema)
	if err != nil {
		return ExecuteResult{}, fmt.Errorf("failed to capture snapshot: %v", err)
	}
	session.putSnapshot(req.Contract, snapshotBytes)

	merges := []base.StateMergeValue(nil)
	if topLevel {
		merges = session.stateMerges()
	}

	if isNewRuntime {
		merges = append(merges, cstate.NewStateMergeValue(
			state.RuntimeStateKey(req.Contract),
			runtimeValue,
		))
	}

	return ExecuteResult{
		Engine:      state.RuntimeEngineGnoSnapshot,
		StateMerges: merges,
	}, nil
}

type normalizedExecuteRequest struct {
	mode             InvocationMode
	registerFunction string
	initData         map[string]string
	callItems        []ExecuteCallItem
}

func normalizeExecuteRequest(req ExecuteRequest) (normalizedExecuteRequest, error) {
	mode := req.Mode
	if mode == "" {
		mode = InvocationModeCall
	}

	switch mode {
	case InvocationModeRegister:
		initData := req.InitData
		if initData == nil {
			initData = req.CallData
		}
		initData = copyCallDataMap(initData)
		if err := ValidateContractCallDataLimits("execute initData", initData); err != nil {
			return normalizedExecuteRequest{}, err
		}

		function := req.Function
		if function == "" {
			function = "Initialize"
		}
		if function != "Initialize" {
			return normalizedExecuteRequest{}, fmt.Errorf("register mode can only execute Initialize, not %q", function)
		}

		return normalizedExecuteRequest{
			mode:             InvocationModeRegister,
			registerFunction: function,
			initData:         initData,
		}, nil
	case InvocationModeCall:
		items, err := normalizeExecuteCallItems(req)
		if err != nil {
			return normalizedExecuteRequest{}, err
		}

		return normalizedExecuteRequest{
			mode:      InvocationModeCall,
			callItems: items,
		}, nil
	default:
		return normalizedExecuteRequest{}, fmt.Errorf("unsupported invocation mode %q", mode)
	}
}

func normalizeExecuteCallItems(req ExecuteRequest) ([]ExecuteCallItem, error) {
	hasLegacy := req.Function != "" || req.CallData != nil
	if len(req.CallItems) > 0 && hasLegacy {
		return nil, fmt.Errorf("call items and legacy function/callData cannot both be set")
	}

	var items []ExecuteCallItem
	switch {
	case len(req.CallItems) > 0:
		items = copyExecuteCallItems(req.CallItems)
	case hasLegacy:
		if err := ValidateContractCallDataLimits("execute callData", req.CallData); err != nil {
			return nil, err
		}
		items = []ExecuteCallItem{
			{
				Function: req.Function,
				CallData: copyCallDataMap(req.CallData),
			},
		}
	default:
		items = nil
	}

	if err := ValidateContractCallItemsLimits("execute call items", items); err != nil {
		return nil, err
	}

	return items, nil
}

func copyExecuteCallItems(items []ExecuteCallItem) []ExecuteCallItem {
	if items == nil {
		return nil
	}

	out := make([]ExecuteCallItem, len(items))
	for i := range items {
		out[i] = ExecuteCallItem{
			Function: items[i].Function,
			CallData: copyCallDataMap(items[i].CallData),
		}
	}

	return out
}

func copyCallDataMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for key, value := range m {
		out[key] = value
	}

	return out
}

func deriveRuntimeState(contract base.Address, source string) state.RuntimeStateValue {
	sum := sha256.Sum256([]byte(contract.String() + ":" + source))
	path := "mitum.local/r/c" + hex.EncodeToString(sum[:8])

	return state.NewRuntimeStateValue(
		state.RuntimeEngineGnoSnapshot,
		string(SchemaModeTypedArgs),
		"contract",
		path,
		GnoSnapshotVersion,
	)
}

func resolveContractSchemaForExecution(preAnalyzed *ContractSchema, sourceCode string) (ContractSchema, error) {
	if preAnalyzed != nil {
		if preAnalyzed.Mode != SchemaModeTypedArgs {
			return ContractSchema{}, fmt.Errorf("gno engine requires typed contract schema")
		}

		return *preAnalyzed, nil
	}

	if schema, found := loadContractSchemaFromCache(sourceCode); found {
		if schema.Mode != SchemaModeTypedArgs {
			return ContractSchema{}, fmt.Errorf("gno engine requires typed contract schema")
		}

		return schema, nil
	}

	schema, err := analyzeContractSchemaFunc(sourceCode)
	if err != nil {
		return ContractSchema{}, err
	}
	if schema.Mode != SchemaModeTypedArgs {
		return ContractSchema{}, fmt.Errorf("Gno engine requires typed contract schema")
	}

	storeContractSchemaInCache(sourceCode, schema)

	return schema, nil
}

func newGnoMachineAndPackage(
	execCtx *ExecutionContext,
	packagePath string,
	contractSource string,
	limits GnoExecutionLimits,
	gasMeter gstore.GasMeter,
) (*gno.Machine, *gno.PackageValue, error) {
	alloc := gno.NewAllocator(limits.MaxAllocBytes)
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})

	store := gno.NewStore(alloc, baseStore, baseStore)
	store.SetNativeResolver(CombineNativeResolvers(MitumNativeResolver, Uint256NativeResolver, gnostdlibs.NativeResolver))

	stdlibPackages, err := GnoStdlibMemPackagesForContract(contractSource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Gno stdlib packages: %w", err)
	}
	purePackages, err := GnoPureMemPackagesForContract(contractSource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Gno pure packages: %w", err)
	}
	runtimePackages := mergeGnoMemPackages(stdlibPackages, purePackages)
	if len(runtimePackages) > 0 {
		// Fixed stdlib and pure-package code is runtime environment setup, not contract work.
		// Persist it to the backing store before opening the metered execution
		// transaction so a permitted import does not spend invocation gas just
		// constructing fixed runtime library packages.
		setupMachine := gno.NewMachineWithOptions(gno.MachineOptions{
			Alloc:              alloc,
			Store:              store,
			Output:             io.Discard,
			Context:            execCtx,
			MaxAllocBytes:      limits.MaxAllocBytes,
			BoundedPanicRender: true,
		})
		for _, rpkg := range runtimePackages {
			if _, _, err := runMemPackage(setupMachine, rpkg, gasMeter); err != nil {
				return nil, nil, fmt.Errorf("failed to load runtime package %q: %w", rpkg.Path, err)
			}
		}
	}

	wrappedStore := baseStore.CacheWrap()
	txStore := store.BeginTransaction(wrappedStore, wrappedStore, nil, gasMeter)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Alloc:              alloc,
		Store:              txStore,
		Output:             io.Discard,
		Context:            execCtx,
		GasMeter:           gasMeter,
		MaxAllocBytes:      limits.MaxAllocBytes,
		BoundedPanicRender: true,
	})

	for _, hpkg := range HostABIMemPackages() {
		hpkg.Type = gno.MPStdlibProd
		if _, _, err := runMemPackage(m, hpkg, gasMeter); err != nil {
			return nil, nil, fmt.Errorf("failed to load host ABI package %q: %w", hpkg.Path, err)
		}
	}

	cpkg := &gnostd.MemPackage{
		Name: "contract",
		Path: packagePath,
		Type: gno.MPUserProd,
		Files: []*gnostd.MemFile{
			{Name: "contract.gno", Body: contractSource},
		},
	}

	_, pv, err := runMemPackage(m, cpkg, gasMeter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load typed contract package: %w", err)
	}

	return m, pv, nil
}

func runMemPackage(m *gno.Machine, pkg *gnostd.MemPackage, gasMeter gstore.GasMeter) (pn *gno.PackageNode, pv *gno.PackageValue, err error) {
	defer func() {
		if r := recover(); r != nil {
			if isGnoPackageLoadResourceLimitPanic(r, gasMeter) {
				panic(r)
			}
			pn = nil
			pv = nil
			err = newGnoPackageLoadPanicError(r)
		}
	}()

	pn, pv = m.RunMemPackage(pkg, true)
	if pn == nil || pv == nil {
		return nil, nil, fmt.Errorf("failed to run mem package %q", pkg.Path)
	}
	return pn, pv, nil
}

func invokeTypedWrite(
	m *gno.Machine,
	pkg *gno.PackageValue,
	req ExecuteRequest,
	schema ContractSchema,
) error {
	fn, found := schema.FindFunction(req.Function)
	if !found {
		return fmt.Errorf("function %q not found", req.Function)
	}

	m.SetActivePackage(pkg)
	m.RunDeclaration(gno.ImportD("chain", MitumChainPackagePath))

	args := []any{
		writeContextExpr(
			req.Sender.String(),
			callerAddressString(req.Caller, req.Sender),
			req.Contract.String(),
			int64(req.Height),
			req.BlockTime,
			false,
		),
	}

	if req.Mode == InvocationModeRegister && fn.Name == "Initialize" {
		initArgs, err := buildInitializeCallArgs(schema, fn, req.CallData)
		if err != nil {
			return err
		}

		args = append(args, initArgs...)
	} else {
		for i := 1; i < len(fn.Params); i++ {
			param := fn.Params[i]
			raw, found := req.CallData[param.Name]
			if !found {
				return fmt.Errorf("missing callData[%q]", param.Name)
			}

			arg, err := buildWriteCallArgExpr(schema, param.Type, raw)
			if err != nil {
				return fmt.Errorf("invalid arg %q: %w", param.Name, err)
			}

			args = append(args, arg)
		}
	}

	results := m.Eval(gno.Call(req.Function, args...))
	if len(results) != 1 {
		return fmt.Errorf("typed write function %q must return exactly one error result", req.Function)
	}

	if !results[0].IsNilInterface() && results[0].IsDefined() {
		if msg := extractTypedWriteErrorMessage(m, results[0]); msg != "" {
			return fmt.Errorf("typed write function %q returned error: %s", req.Function, msg)
		}

		return fmt.Errorf("typed write function %q returned non-nil error", req.Function)
	}

	return nil
}

func buildInitializeCallArgs(schema ContractSchema, fn FunctionSchema, callData map[string]string) ([]any, error) {
	expected := map[string]ParamSchema{}
	for i := 1; i < len(fn.Params); i++ {
		expected[fn.Params[i].Name] = fn.Params[i]
	}

	for key := range callData {
		if _, found := expected[key]; !found {
			return nil, fmt.Errorf(`unknown initialize arg %q`, key)
		}
	}

	args := make([]any, 0, len(fn.Params)-1)
	for i := 1; i < len(fn.Params); i++ {
		param := fn.Params[i]
		raw, found := callData[param.Name]
		if !found {
			return nil, fmt.Errorf(`missing required initialize arg %q`, param.Name)
		}

		arg, err := buildInitializeCallArgExpr(schema, param, raw)
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
	}

	return args, nil
}

func extractTypedWriteErrorMessage(m *gno.Machine, tv gno.TypedValue) string {
	msg := strings.TrimSpace(safeTypedValueSprint(m, tv))
	if msg != "" && msg != "undefined" {
		return msg
	}

	msg = strings.TrimSpace(safeTypedValueString(tv))
	if msg != "" && msg != "(undefined)" {
		return msg
	}

	return ""
}

func safeTypedValueSprint(m *gno.Machine, tv gno.TypedValue) (out string) {
	defer func() {
		if recover() != nil {
			out = ""
		}
	}()

	return tv.Sprint(m)
}

func safeTypedValueString(tv gno.TypedValue) (out string) {
	defer func() {
		if recover() != nil {
			out = ""
		}
	}()

	return tv.String()
}

func callerAddressString(caller base.Address, sender base.Address) string {
	if caller != nil {
		return caller.String()
	}

	return sender.String()
}

func writeContextExpr(sender, caller, contract string, height int64, blockTime int64, readOnly bool) gno.Expr {
	return &gno.CompositeLitExpr{
		Type: gno.Sel(gno.Nx("chain"), "WriteContext"),
		Elts: gno.KeyValueExprs{
			gno.Kv("Sender", gno.Str(sender)),
			gno.Kv("Caller", gno.Str(caller)),
			gno.Kv("Contract", gno.Str(contract)),
			gno.Kv("Height", gno.Num(strconv.FormatInt(height, 10))),
			gno.Kv("BlockTime", gno.Num(strconv.FormatInt(blockTime, 10))),
			gno.Kv("ReadOnly", gno.X(strconv.FormatBool(readOnly))),
		},
	}
}

func queryContextExpr(contract string, height int64, currentHeight int64, readOnly bool) gno.Expr {
	return &gno.CompositeLitExpr{
		Type: gno.Sel(gno.Nx("chain"), "QueryContext"),
		Elts: gno.KeyValueExprs{
			gno.Kv("Contract", gno.Str(contract)),
			gno.Kv("Height", gno.Num(strconv.FormatInt(height, 10))),
			gno.Kv("CurrentHeight", gno.Num(strconv.FormatInt(currentHeight, 10))),
			gno.Kv("ReadOnly", gno.X(strconv.FormatBool(readOnly))),
		},
	}
}
