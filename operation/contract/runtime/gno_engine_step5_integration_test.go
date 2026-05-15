package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const mapScalarEngineContractSource = `package contract
import "mitum/chain"

var balances map[string]int64

func Initialize(ctx chain.ContractContext) error {
	balances = map[string]int64{"alice":1}
	return nil
}

func AddBalance(ctx chain.ContractContext, owner string, amount int64) error {
	if balances == nil {
		balances = map[string]int64{}
	}
	balances[owner] = balances[owner] + amount
	return nil
}
`

const mapStructEngineContractSource = `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error {
	users = map[string]User{
		"alice": User{Balance:1, Active:true},
	}
	return nil
}

func UpdateUser(ctx chain.ContractContext, owner string, balance int64, active bool) error {
	if users == nil {
		users = map[string]User{}
	}
	users[owner] = User{Balance:balance, Active:active}
	return nil
}
`

func TestGnoWritePathMapStringScalarRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(mapScalarEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0002")
	sender := base.NewStringAddress("sender0002")
	runtimeValue := deriveRuntimeState(contract, mapScalarEngineContractSource)
	getStateFunc := func(key string) (base.State, bool, error) {
		return nil, false, nil
	}

	registerSnapshot := executeGnoWriteForTest(
		t,
		encs,
		getStateFunc,
		schema,
		runtimeValue.PackagePath,
		ExecuteRequest{
			Mode:         InvocationModeRegister,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(20),
			ContractCode: mapScalarEngineContractSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		},
		nil,
	)
	assertMapScalarBinding(t, registerSnapshot, "balances", map[string]string{
		"alice": "1",
	})

	callSnapshot := executeGnoWriteForTest(
		t,
		encs,
		getStateFunc,
		schema,
		runtimeValue.PackagePath,
		ExecuteRequest{
			Mode:         InvocationModeCall,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(21),
			ContractCode: mapScalarEngineContractSource,
			Function:     "AddBalance",
			CallData: map[string]string{
				"owner":  "bob",
				"amount": "7",
			},
		},
		mustMarshalSnapshotDoc(t, registerSnapshot),
	)
	assertMapScalarBinding(t, callSnapshot, "balances", map[string]string{
		"alice": "1",
		"bob":   "7",
	})
}

func TestGnoWritePathMapStringStructRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(mapStructEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0003")
	sender := base.NewStringAddress("sender0003")
	runtimeValue := deriveRuntimeState(contract, mapStructEngineContractSource)
	getStateFunc := func(key string) (base.State, bool, error) {
		return nil, false, nil
	}

	registerSnapshot := executeGnoWriteForTest(
		t,
		encs,
		getStateFunc,
		schema,
		runtimeValue.PackagePath,
		ExecuteRequest{
			Mode:         InvocationModeRegister,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(30),
			ContractCode: mapStructEngineContractSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		},
		nil,
	)
	assertMapStructBinding(t, registerSnapshot, "users", map[string]map[string]string{
		"alice": {
			"Balance": "1",
			"Active":  "true",
		},
	})

	callSnapshot := executeGnoWriteForTest(
		t,
		encs,
		getStateFunc,
		schema,
		runtimeValue.PackagePath,
		ExecuteRequest{
			Mode:         InvocationModeCall,
			Contract:     contract,
			Sender:       sender,
			Height:       base.Height(31),
			ContractCode: mapStructEngineContractSource,
			Function:     "UpdateUser",
			CallData: map[string]string{
				"owner":   "bob",
				"balance": "7",
				"active":  "false",
			},
		},
		mustMarshalSnapshotDoc(t, registerSnapshot),
	)
	assertMapStructBinding(t, callSnapshot, "users", map[string]map[string]string{
		"alice": {
			"Balance": "1",
			"Active":  "true",
		},
		"bob": {
			"Balance": "7",
			"Active":  "false",
		},
	})
}

func assertMapScalarBinding(t *testing.T, doc SnapshotDoc, name string, expected map[string]string) {
	t.Helper()

	binding := findSnapshotBinding(t, doc, name)
	if binding.Value.Kind != string(TypeMap) {
		t.Fatalf("expected map snapshot binding, got %#v", binding.Value)
	}
	if binding.Value.IsNil {
		t.Fatalf("expected non-nil map snapshot")
	}
	if len(binding.Value.Entries) != len(expected) {
		t.Fatalf("unexpected entry count: %#v", binding.Value.Entries)
	}

	for _, entry := range binding.Value.Entries {
		want, found := expected[entry.Key]
		if !found {
			t.Fatalf("unexpected entry key %q", entry.Key)
		}
		if entry.Value.Kind != string(TypeScalar) || entry.Value.Scalar != want {
			t.Fatalf("unexpected scalar entry for %q: %#v", entry.Key, entry.Value)
		}
	}
}

func assertMapStructBinding(t *testing.T, doc SnapshotDoc, name string, expected map[string]map[string]string) {
	t.Helper()

	binding := findSnapshotBinding(t, doc, name)
	if binding.Value.Kind != string(TypeMap) {
		t.Fatalf("expected map snapshot binding, got %#v", binding.Value)
	}
	if binding.Value.IsNil {
		t.Fatalf("expected non-nil map snapshot")
	}
	if len(binding.Value.Entries) != len(expected) {
		t.Fatalf("unexpected entry count: %#v", binding.Value.Entries)
	}

	for _, entry := range binding.Value.Entries {
		wantFields, found := expected[entry.Key]
		if !found {
			t.Fatalf("unexpected entry key %q", entry.Key)
		}
		if entry.Value.Kind != string(TypeStruct) {
			t.Fatalf("expected struct entry value for %q, got %#v", entry.Key, entry.Value)
		}
		if len(entry.Value.Fields) != len(wantFields) {
			t.Fatalf("unexpected field count for %q: %#v", entry.Key, entry.Value.Fields)
		}

		for _, field := range entry.Value.Fields {
			want, found := wantFields[field.Name]
			if !found {
				t.Fatalf("unexpected struct field %q for %q", field.Name, entry.Key)
			}
			if field.Value.Kind != string(TypeScalar) || field.Value.Scalar != want {
				t.Fatalf("unexpected field value for %q.%s: %#v", entry.Key, field.Name, field.Value)
			}
		}
	}
}

func findSnapshotBinding(t *testing.T, doc SnapshotDoc, name string) SnapshotBinding {
	t.Helper()

	for _, binding := range doc.Bindings {
		if binding.Name == name {
			return binding
		}
	}

	t.Fatalf("binding %q not found in snapshot doc", name)
	return SnapshotBinding{}
}
