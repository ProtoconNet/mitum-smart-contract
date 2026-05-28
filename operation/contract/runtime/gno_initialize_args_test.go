package runtime

import (
	"strings"
	"testing"

	pstate "github.com/ProtoconNet/mitum-smart-contract/state"
	"github.com/ProtoconNet/mitum2/base"
)

const initializeArgsContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Limit int64
}

type Record struct {
	Name string
	Meta Meta
}

var seedOwner string
var seedLabel string
var seedLimit int64
var record Record

func Initialize(ctx chain.WriteContext, owner string, label string, limit int64) error {
	seedOwner = owner
	seedLabel = label
	seedLimit = limit
	record = Record{
		Name: label,
		Meta: Meta{
			Limit: limit,
		},
	}
	return nil
}

func GetOwner(ctx chain.QueryContext) string { return seedOwner }
func GetLabel(ctx chain.QueryContext) string { return seedLabel }
func GetLimit(ctx chain.QueryContext) int64 { return seedLimit }
func GetRecord(ctx chain.QueryContext) Record { return record }
`

const initializeNoArgsContractSource = `package contract
import "mitum/chain"

var initialized bool

func Initialize(ctx chain.WriteContext) error {
	initialized = true
	return nil
}

func IsInitialized(ctx chain.QueryContext) bool { return initialized }
`

func TestInitializeArgsRegisterSuccess(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractinit0001")
	sender := base.NewStringAddress("senderinit0001")

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(600),
		ContractCode: initializeArgsContractSource,
		Function:     "Initialize",
		CallData: map[string]string{
			"owner": "alice",
			"label": "demo",
			"limit": "10",
		},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(600), result.StateMerges)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: initializeArgsContractSource,
		Function:     "GetOwner",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetOwner) returned error: %v", err)
	}
	if got := qr.Result.(string); got != "alice" {
		t.Fatalf("expected owner alice, got %#v", qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: initializeArgsContractSource,
		Function:     "GetLimit",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetLimit) returned error: %v", err)
	}
	if got := qr.Result.(int64); got != 10 {
		t.Fatalf("expected limit 10, got %#v", qr.Result)
	}

	qr, err = engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: initializeArgsContractSource,
		Function:     "GetRecord",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(GetRecord) returned error: %v", err)
	}
	assertDeepEqualResult(t, qr.Result, map[string]interface{}{
		"Name": "demo",
		"Meta": map[string]interface{}{
			"Limit": int64(10),
		},
	})
}

func TestInitializeArgsMissingRequiredArgFails(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractinit0002")
	sender := base.NewStringAddress("senderinit0002")

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(600),
		ContractCode: initializeArgsContractSource,
		Function:     "Initialize",
		CallData: map[string]string{
			"owner": "alice",
			"label": "demo",
		},
	})
	if err == nil {
		t.Fatalf("expected missing initialize arg error")
	}
	if !strings.Contains(err.Error(), `missing required initialize arg "limit"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitializeArgsUnknownKeyFails(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractinit0003")
	sender := base.NewStringAddress("senderinit0003")

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(600),
		ContractCode: initializeArgsContractSource,
		Function:     "Initialize",
		CallData: map[string]string{
			"owner": "alice",
			"label": "demo",
			"limit": "10",
			"extra": "oops",
		},
	})
	if err == nil {
		t.Fatalf("expected unknown initialize arg error")
	}
	if !strings.Contains(err.Error(), `unknown initialize arg "extra"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitializeArgsParseFailureFails(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractinit0004")
	sender := base.NewStringAddress("senderinit0004")

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(600),
		ContractCode: initializeArgsContractSource,
		Function:     "Initialize",
		CallData: map[string]string{
			"owner": "alice",
			"label": "demo",
			"limit": "not-a-number",
		},
	})
	if err == nil {
		t.Fatalf("expected initialize arg parse error")
	}
	if !containsAll(err.Error(), `failed to parse initialize arg "limit" as int64`, "invalid syntax") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitializeNoArgsRegisterStillSupported(t *testing.T) {
	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractinit0005")
	sender := base.NewStringAddress("senderinit0005")

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), func(string) (base.State, bool, error) {
		return nil, false, nil
	}, ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(600),
		ContractCode: initializeNoArgsContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}

	states := map[string]base.State{}
	applyStateMerges(states, base.Height(600), result.StateMerges)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       states[pstate.SnapshotStateKey(contract)].Height(),
		ContractCode: initializeNoArgsContractSource,
		Function:     "IsInitialized",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract(IsInitialized) returned error: %v", err)
	}
	if got := qr.Result.(bool); !got {
		t.Fatalf("expected initialized=true")
	}
}
