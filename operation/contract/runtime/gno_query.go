package runtime

import (
	"bytes"
	"fmt"

	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gstore "github.com/gnolang/gno/tm2/pkg/store"
)

func (gnoEngine) QueryContract(
	encs encoder.Encoders,
	getStateFunc base.GetStateFunc,
	req QueryRequest,
) (_ QueryResult, berr base.OperationProcessReasonError) {
	var gasMeter gstore.GasMeter

	defer func() {
		if r := recover(); r != nil {
			berr = ClassifyGnoExecutionPanic("gno query", r, gasMeter)
		}
	}()

	if err := ValidateContractCallDataLimits("query callData", req.CallData); err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("invalid call data: %v", err)
	}

	schema, err := resolveContractSchemaForExecution(req.Schema, req.ContractCode)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to analyze contract schema: %v", err)
	}

	fn, found := schema.FindFunction(req.Function)
	if !found {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("query function %q not found", req.Function)
	}
	if !fn.IsTypedQueryShape() {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("function %q is not a supported query shape", req.Function)
	}

	rst, found, err := getStateFunc(pstate.RuntimeStateKey(req.Contract))
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to read runtime state: %v", err)
	}
	if !found {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("runtime state not found for typed contract %v", req.Contract)
	}

	runtimeValue, err := pstate.GetRuntimeFromState(rst)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to decode runtime state: %v", err)
	}
	if runtimeValue.Engine != pstate.RuntimeEngineGnoSnapshot {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("runtime engine mismatch: %q", runtimeValue.Engine)
	}

	sst, found, err := getStateFunc(pstate.SnapshotStateKey(req.Contract))
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to read snapshot state: %v", err)
	}
	if !found {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("snapshot state not found for typed contract %v", req.Contract)
	}

	snapshotValue, err := pstate.GetSnapshotFromState(sst)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to decode snapshot state: %v", err)
	}
	if snapshotValue.Version != GnoSnapshotVersion {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("unsupported snapshot version %d", snapshotValue.Version)
	}
	if snapshotValue.Codec != GnoSnapshotCodecName {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("unsupported snapshot codec %q", snapshotValue.Codec)
	}

	execCtx, err := NewExecutionContext(
		encs,
		getStateFunc,
		req.Contract,
		req.Sender,
		req.Height,
		true,
	)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to build execution context: %v", err)
	}

	limits := QueryGnoExecutionLimits()
	gasMeter = NewGnoGasMeter(limits.GasLimit)

	m, pkg, err := newGnoMachineAndPackage(
		execCtx,
		runtimeValue.PackagePath,
		req.ContractCode,
		limits,
		gasMeter,
	)

	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to initialize gno machine: %v", err)
	}

	if err := RestoreSnapshot(m, pkg, snapshotValue.Snapshot, schema); err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to restore snapshot: %v", err)
	}

	before, err := CaptureSnapshot(pkg, m.Store, schema)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to capture pre-query snapshot: %v", err)
	}

	qr, err := invokeTypedQuery(m, pkg, req, schema)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to execute query: %v", err)
	}

	after, err := CaptureSnapshot(pkg, m.Store, schema)
	if err != nil {
		return QueryResult{}, base.NewBaseOperationProcessReasonError("failed to capture post-query snapshot: %v", err)
	}
	if !bytes.Equal(before, after) {
		return QueryResult{}, base.NewBaseOperationProcessReasonError(
			"query function %q mutated contract state", req.Function,
		)
	}

	qr.Engine = pstate.RuntimeEngineGnoSnapshot
	return qr, nil
}

func invokeTypedQuery(
	m *gno.Machine,
	pkg *gno.PackageValue,
	req QueryRequest,
	schema ContractSchema,
) (QueryResult, error) {
	fn, found := schema.FindFunction(req.Function)
	if !found {
		return QueryResult{}, fmt.Errorf("function %q not found", req.Function)
	}

	m.SetActivePackage(pkg)
	m.RunDeclaration(gno.ImportD("chain", MitumChainPackagePath))

	args := []any{
		contractContextExpr(req.Sender.String(), req.Contract.String(), int64(req.Height), true),
	}

	for i := 1; i < len(fn.Params); i++ {
		param := fn.Params[i]
		raw, found := req.CallData[param.Name]
		if !found {
			return QueryResult{}, fmt.Errorf("missing callData[%q]", param.Name)
		}

		arg, err := buildQueryCallArgExpr(schema, param.Type, raw)
		if err != nil {
			return QueryResult{}, fmt.Errorf("invalid arg %q: %w", param.Name, err)
		}

		args = append(args, arg)
	}

	results := m.Eval(gno.Call(req.Function, args...))

	switch len(fn.Results) {
	case 1:
		if len(results) != 1 {
			return QueryResult{}, fmt.Errorf("query function %q must return exactly 1 value", req.Function)
		}

		qv, err := ExtractQueryValue(schema, fn.Results[0].Type, results[0], m.Store)
		if err != nil {
			return QueryResult{}, err
		}
		v, err := QueryValueToJSONCompatible(schema, fn.Results[0].Type, qv)
		if err != nil {
			return QueryResult{}, err
		}

		return QueryResult{Result: v}, nil

	case 2:
		if len(results) != 2 {
			return QueryResult{}, fmt.Errorf("query function %q must return exactly 2 values", req.Function)
		}

		qv, err := ExtractQueryValue(schema, fn.Results[0].Type, results[0], m.Store)
		if err != nil {
			return QueryResult{}, err
		}
		v, err := QueryValueToJSONCompatible(schema, fn.Results[0].Type, qv)
		if err != nil {
			return QueryResult{}, err
		}

		ok := results[1].GetBool()
		return QueryResult{
			Result: v,
			Ok:     &ok,
		}, nil

	default:
		return QueryResult{}, fmt.Errorf("unsupported query result count for %q", req.Function)
	}
}

func ParseOptionalQuerySender(
	encs encoder.Encoders,
	contract base.Address,
	callData map[string]string,
) (base.Address, error) {
	raw, found := callData["_sender"]
	if !found || raw == "" {
		return contract, nil
	}

	sender, err := base.DecodeAddress(raw, encs.JSON())
	if err != nil {
		return nil, fmt.Errorf("invalid _sender: %w", err)
	}

	return sender, nil
}
