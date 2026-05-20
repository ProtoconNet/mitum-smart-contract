package digest

import (
	"net/http"
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum-currency/v3/common"
	cruntime "github.com/ProtoconNet/mitum-currency/v3/operation/contract/runtime"
	pstate "github.com/ProtoconNet/mitum-currency/v3/state/contract"
	ptypes "github.com/ProtoconNet/mitum-currency/v3/types/contract"
	"github.com/ProtoconNet/mitum2/base"
	"github.com/ProtoconNet/mitum2/util/encoder"
)

type digestQuerySpyEngine struct {
	queryCalls int
}

func (e *digestQuerySpyEngine) ValidateContract(string) (cruntime.ContractSchema, base.OperationProcessReasonError) {
	return cruntime.ContractSchema{}, nil
}

func (e *digestQuerySpyEngine) ExecuteContract(
	encoder.Encoders,
	base.GetStateFunc,
	cruntime.ExecuteRequest,
) (cruntime.ExecuteResult, base.OperationProcessReasonError) {
	return cruntime.ExecuteResult{}, nil
}

func (e *digestQuerySpyEngine) QueryContract(
	encoder.Encoders,
	base.GetStateFunc,
	cruntime.QueryRequest,
) (cruntime.QueryResult, base.OperationProcessReasonError) {
	e.queryCalls++
	return cruntime.QueryResult{}, nil
}

func TestContractQueryEndpointMalformedJSONStopsBeforeRuntimeQuery(t *testing.T) {
	spy := useDigestQuerySpyEngine(t)

	hd, contract, _ := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{Mode: cruntime.InvocationModeRegister, Height: base.Height(700), Function: "Initialize", CallData: map[string]string{}},
	})

	status, body, _ := performRawContractQueryRequest(t, hd, contract, `{"function":"GetOwner"`)
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if spy.queryCalls != 0 {
		t.Fatalf("expected runtime query not to be called, got %d calls", spy.queryCalls)
	}
}

func TestContractQueryEndpointRawBodyLimitStopsBeforeRuntimeQuery(t *testing.T) {
	spy := useDigestQuerySpyEngine(t)

	hd, contract, _ := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{Mode: cruntime.InvocationModeRegister, Height: base.Height(701), Function: "Initialize", CallData: map[string]string{}},
	})

	status, body, _ := performRawContractQueryRequest(t, hd, contract, strings.Repeat("{", MaxContractQueryBodyBytes+1))
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if !strings.Contains(body, "query body exceeds max size") {
		t.Fatalf("unexpected body: %s", body)
	}
	if spy.queryCalls != 0 {
		t.Fatalf("expected runtime query not to be called, got %d calls", spy.queryCalls)
	}
}

func TestContractQueryEndpointDecodedCallDataLimitStopsBeforeRuntimeQuery(t *testing.T) {
	spy := useDigestQuerySpyEngine(t)

	hd, contract, _ := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{Mode: cruntime.InvocationModeRegister, Height: base.Height(702), Function: "Initialize", CallData: map[string]string{}},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, digestQueryPayloadEntries(cruntime.MaxContractCallDataEntries+1))
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if !strings.Contains(body, "max entries") {
		t.Fatalf("unexpected body: %s", body)
	}
	if spy.queryCalls != 0 {
		t.Fatalf("expected runtime query not to be called, got %d calls", spy.queryCalls)
	}
}

func TestContractQueryEndpointMissingFunctionStopsBeforeRuntimeQuery(t *testing.T) {
	spy := useDigestQuerySpyEngine(t)

	hd, contract, _ := newTypedQueryTestHandlers(t, typedDigestQueryContractSource, []cruntime.ExecuteRequest{
		{Mode: cruntime.InvocationModeRegister, Height: base.Height(703), Function: "Initialize", CallData: map[string]string{}},
	})

	status, body, _ := performContractQueryRequest(t, hd, contract, map[string]string{"name": "alice"})
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if !strings.Contains(body, "missing function in query body") {
		t.Fatalf("unexpected body: %s", body)
	}
	if spy.queryCalls != 0 {
		t.Fatalf("expected runtime query not to be called, got %d calls", spy.queryCalls)
	}
}

func TestContractQueryEndpointContractNotFoundStopsBeforeRuntimeQuery(t *testing.T) {
	spy := useDigestQuerySpyEngine(t)

	encs, enc := newDigestTestEncoders(t)
	hd := newDigestHandlersForStates(t, encs, enc, map[string]base.State{})

	status, body, _ := performContractQueryRequest(t, hd, base.NewStringAddress("contractdnf0001").String(), map[string]string{
		"function": "GetOwner",
	})
	if status != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if !strings.Contains(body, "contract design not found") {
		t.Fatalf("unexpected body: %s", body)
	}
	if spy.queryCalls != 0 {
		t.Fatalf("expected runtime query not to be called, got %d calls", spy.queryCalls)
	}
}

func TestContractQueryEndpointSnapshotStateMissingStopsBeforeRuntimeQuery(t *testing.T) {
	spy := useDigestQuerySpyEngine(t)

	encs, enc := newDigestTestEncoders(t)
	contract := base.NewStringAddress("contractsnp0001")
	states := map[string]base.State{
		pstate.DesignStateKey(contract): common.NewBaseState(
			base.Height(710),
			pstate.DesignStateKey(contract),
			pstate.NewDesignStateValue(ptypes.NewDesign(typedDigestQueryContractSource)),
			nil,
			nil,
		),
		pstate.RuntimeStateKey(contract): common.NewBaseState(
			base.Height(710),
			pstate.RuntimeStateKey(contract),
			pstate.NewRuntimeStateValue(
				pstate.RuntimeEngineGnoSnapshot,
				string(cruntime.SchemaModeTypedArgs),
				"contract",
				"mitum.local/r/csnapshot",
				cruntime.GnoSnapshotVersion,
			),
			nil,
			nil,
		),
	}
	hd := newDigestHandlersForStates(t, encs, enc, states)

	status, body, _ := performContractQueryRequest(t, hd, contract.String(), map[string]string{
		"function": "GetOwner",
	})
	if status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if !strings.Contains(body, "snapshot state not found") {
		t.Fatalf("unexpected body: %s", body)
	}
	if spy.queryCalls != 0 {
		t.Fatalf("expected runtime query not to be called, got %d calls", spy.queryCalls)
	}
}

func TestContractQueryEndpointRuntimeStateMissingReturnsConsistentSignal(t *testing.T) {
	encs, enc := newDigestTestEncoders(t)
	contract := base.NewStringAddress("contractrtm0001")
	states := map[string]base.State{
		pstate.DesignStateKey(contract): common.NewBaseState(
			base.Height(711),
			pstate.DesignStateKey(contract),
			pstate.NewDesignStateValue(ptypes.NewDesign(typedDigestQueryContractSource)),
			nil,
			nil,
		),
	}
	hd := newDigestHandlersForStates(t, encs, enc, states)

	status, body, _ := performContractQueryRequest(t, hd, contract.String(), map[string]string{
		"function": "GetOwner",
	})
	if status != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d body=%s", status, body)
	}
	if !strings.Contains(body, "runtime state not found for typed contract") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func useDigestQuerySpyEngine(t *testing.T) *digestQuerySpyEngine {
	t.Helper()

	original := digestContractQueryEngine
	spy := &digestQuerySpyEngine{}
	digestContractQueryEngine = spy
	t.Cleanup(func() {
		digestContractQueryEngine = original
	})

	return spy
}
