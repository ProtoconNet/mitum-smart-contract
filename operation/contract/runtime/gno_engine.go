package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	"github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnostdlibs "github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

type gnoEngine struct{}

var analyzeContractSchemaFunc = AnalyzeContractSchema
var newGnoMachineAndPackageFunc = newGnoMachineAndPackage

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

	if err := ValidateContractCallDataLimits("execute callData", req.CallData); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("invalid call data: %v", err)
	}

	schema, err := resolveContractSchemaForExecution(req.Schema, req.ContractCode)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	var runtimeValue state.RuntimeStateValue
	isNewRuntime := false

	if st, found, err := getStateFunc(state.RuntimeStateKey(req.Contract)); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	} else if found {
		runtimeValue, err = state.GetRuntimeFromState(st)
		if err != nil {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
		}
		if runtimeValue.Engine != state.RuntimeEngineGnoSnapshot {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
				"runtime engine mismatch: %q", runtimeValue.Engine,
			)
		}
	} else {
		if req.Mode != InvocationModeRegister {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
				"runtime state not found for typed contract %v; register must create runtime first",
				req.Contract,
			)
		}

		runtimeValue = deriveRuntimeState(req.Contract, req.ContractCode)
		isNewRuntime = true
	}

	var snapshotValue state.SnapshotStateValue

	if st, found, err := getStateFunc(state.SnapshotStateKey(req.Contract)); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to read snapshot state: %v", err)
	} else if found {
		snapshotValue, err = state.GetSnapshotFromState(st)
		if err != nil {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to decode snapshot state: %v", err)
		}
	} else {
		if req.Mode != InvocationModeRegister {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
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
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
			"unsupported snapshot version %d", snapshotValue.Version,
		)
	}
	if snapshotValue.Codec != GnoSnapshotCodecName {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
			"unsupported snapshot codec %q", snapshotValue.Codec,
		)
	}

	execCtx, err := NewExecutionContext(
		encs,
		getStateFunc,
		req.Contract,
		req.Sender,
		req.Height,
		false,
	)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to build execution context: %v", err)
	}

	limits := WriteGnoExecutionLimits()
	gasMeter = NewGnoGasMeter(limits.GasLimit)

	m, pkg, err := newGnoMachineAndPackageFunc(
		execCtx,
		runtimeValue.PackagePath,
		req.ContractCode,
		limits,
		gasMeter,
	)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to initialize gno machine: %v", err)
	}

	if err := RestoreSnapshot(m, pkg, snapshotValue.Snapshot, schema); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to restore snapshot: %v", err)
	}

	if req.Mode == InvocationModeCall && req.Function == "Initialize" {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError(
			"Initialize cannot be called through call operation for typed contracts",
		)
	}

	if err := invokeTypedWrite(m, pkg, req, schema); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to execute typed contract call: %v", err)
	}

	snapshotBytes, err := CaptureSnapshot(pkg, m.Store, schema)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to capture snapshot: %v", err)
	}

	merges := []base.StateMergeValue{
		cstate.NewStateMergeValue(
			state.SnapshotStateKey(req.Contract),
			state.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, snapshotBytes),
		),
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
	store.SetNativeResolver(CombineNativeResolvers(MitumNativeResolver, gnostdlibs.NativeResolver))

	stdlibPackages, err := GnoStdlibMemPackagesForContract(contractSource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Gno stdlib packages: %w", err)
	}
	if len(stdlibPackages) > 0 {
		// Loaded stdlib code is runtime environment setup, not contract work.
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
		for _, spkg := range stdlibPackages {
			if _, _, err := runMemPackage(setupMachine, spkg, gasMeter); err != nil {
				return nil, nil, fmt.Errorf("failed to load stdlib package %q: %w", spkg.Path, err)
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
		writeContextExpr(req.Sender.String(), req.Contract.String(), int64(req.Height), req.BlockTime, false),
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

func writeContextExpr(sender, contract string, height int64, blockTime int64, readOnly bool) gno.Expr {
	return &gno.CompositeLitExpr{
		Type: gno.Sel(gno.Nx("chain"), "WriteContext"),
		Elts: gno.KeyValueExprs{
			gno.Kv("Sender", gno.Str(sender)),
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
