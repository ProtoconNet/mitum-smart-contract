package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	cstate "github.com/ProtoconNet/mitum-currency/v3/state"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	gnostd "github.com/gnolang/gno/tm2/pkg/std"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

type gnoEngine struct{}

func NewGnoEngine() ContractEngine {
	return gnoEngine{}
}

func (gnoEngine) ValidateContract(sourceCode string) base.OperationProcessReasonError {
	schema, err := AnalyzeContractSchema(sourceCode)
	if err != nil {
		return base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}
	if schema.Mode != SchemaModeTypedArgs {
		return base.NewBaseOperationProcessReasonError("Gno engine requires typed contract schema")
	}

	return nil
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

	schema, err := AnalyzeContractSchema(req.ContractCode)
	if err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	var runtimeValue pstate.RuntimeStateValue
	isNewRuntime := false

	if st, found, err := getStateFunc(pstate.RuntimeStateKey(req.Contract)); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	} else if found {
		runtimeValue, err = pstate.GetRuntimeFromState(st)
		if err != nil {
			return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
		}
		if runtimeValue.Engine != pstate.RuntimeEngineGnoSnapshot {
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

	var snapshotValue pstate.SnapshotStateValue

	if st, found, err := getStateFunc(pstate.SnapshotStateKey(req.Contract)); err != nil {
		return ExecuteResult{}, base.NewBaseOperationProcessReasonError("failed to read snapshot state: %v", err)
	} else if found {
		snapshotValue, err = pstate.GetSnapshotFromState(st)
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

		snapshotValue = pstate.NewSnapshotStateValue(
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

	m, pkg, err := newGnoMachineAndPackage(
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
			pstate.SnapshotStateKey(req.Contract),
			pstate.NewSnapshotStateValue(GnoSnapshotVersion, GnoSnapshotCodecName, snapshotBytes),
		),
	}

	if isNewRuntime {
		merges = append(merges, cstate.NewStateMergeValue(
			pstate.RuntimeStateKey(req.Contract),
			runtimeValue,
		))
	}

	return ExecuteResult{
		Engine:      pstate.RuntimeEngineGnoSnapshot,
		StateMerges: merges,
	}, nil
}

func deriveRuntimeState(contract base.Address, source string) pstate.RuntimeStateValue {
	sum := sha256.Sum256([]byte(contract.String() + ":" + source))
	path := "mitum.local/r/c" + hex.EncodeToString(sum[:8])

	return pstate.NewRuntimeStateValue(
		pstate.RuntimeEngineGnoSnapshot,
		string(SchemaModeTypedArgs),
		"contract",
		path,
		GnoSnapshotVersion,
	)
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
	wrappedStore := baseStore.CacheWrap()

	store := gno.NewStore(alloc, baseStore, baseStore)
	store.SetNativeResolver(CombineNativeResolvers(MitumNativeResolver))
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
		if _, _, err := runMemPackage(m, hpkg); err != nil {
			return nil, nil, err
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

	_, pv, err := runMemPackage(m, cpkg)
	if err != nil {
		return nil, nil, err
	}

	return m, pv, nil
}

func runMemPackage(m *gno.Machine, pkg *gnostd.MemPackage) (*gno.PackageNode, *gno.PackageValue, error) {
	pn, pv := m.RunMemPackage(pkg, true)
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
		contractContextExpr(req.Sender.String(), req.Contract.String(), int64(req.Height), false),
	}

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

func contractContextExpr(sender, contract string, height int64, readOnly bool) gno.Expr {
	return &gno.CompositeLitExpr{
		Type: gno.Sel(gno.Nx("chain"), "ContractContext"),
		Elts: gno.KeyValueExprs{
			gno.Kv("Sender", gno.Str(sender)),
			gno.Kv("Contract", gno.Str(contract)),
			gno.Kv("Height", gno.Num(strconv.FormatInt(height, 10))),
			gno.Kv("ReadOnly", gno.X(strconv.FormatBool(readOnly))),
		},
	}
}
