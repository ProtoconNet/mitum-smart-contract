package runtime

import (
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const nestedStructEngineContractSource = `package contract
import "mitum/chain"

type Limits struct {
	Daily int64
	Max int64
}

type Config struct {
	Owner string
	Limits Limits
}

var config Config

func Initialize(ctx chain.WriteContext) error {
	config.Owner = ctx.GetSender()
	config.Limits.Daily = 1
	config.Limits.Max = 2
	return nil
}

func UpdateLimits(ctx chain.WriteContext, daily int64, max int64) error {
	config.Limits.Daily = daily
	config.Limits.Max = max
	return nil
}
`

const mapNestedStructEngineContractSource = `package contract
import "mitum/chain"

type Meta struct {
	Active bool
	Limit int64
}

type User struct {
	Balance int64
	Meta Meta
}

var users map[string]User

func Initialize(ctx chain.WriteContext) error {
	users = map[string]User{
		"alice": User{Balance:1, Meta:Meta{Active:true, Limit:10}},
	}
	return nil
}

func UpdateUser(ctx chain.WriteContext, owner string, balance int64, active bool, limit int64) error {
	if users == nil {
		users = map[string]User{}
	}
	users[owner] = User{Balance:balance, Meta:Meta{Active:active, Limit:limit}}
	return nil
}
`

func TestGnoWritePathNestedStructRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(nestedStructEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0004")
	sender := base.NewStringAddress("sender0004")
	runtimeValue := deriveRuntimeState(contract, nestedStructEngineContractSource)
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
			Height:       base.Height(90),
			ContractCode: nestedStructEngineContractSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		},
		nil,
	)
	assertNestedStructBinding(t, registerSnapshot, "config", map[string]interface{}{
		"Owner": sender.String(),
		"Limits": map[string]interface{}{
			"Daily": "1",
			"Max":   "2",
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
			Height:       base.Height(91),
			ContractCode: nestedStructEngineContractSource,
			Function:     "UpdateLimits",
			CallData: map[string]string{
				"daily": "7",
				"max":   "11",
			},
		},
		mustMarshalSnapshotDoc(t, registerSnapshot),
	)
	assertNestedStructBinding(t, callSnapshot, "config", map[string]interface{}{
		"Owner": sender.String(),
		"Limits": map[string]interface{}{
			"Daily": "7",
			"Max":   "11",
		},
	})
}

func TestGnoWritePathMapStringNestedStructRoundTrip(t *testing.T) {
	encs := newRuntimeTestEncoders(t)
	schema, err := AnalyzeContractSchema(mapNestedStructEngineContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	contract := base.NewStringAddress("contract0005")
	sender := base.NewStringAddress("sender0005")
	runtimeValue := deriveRuntimeState(contract, mapNestedStructEngineContractSource)
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
			Height:       base.Height(100),
			ContractCode: mapNestedStructEngineContractSource,
			Function:     "Initialize",
			CallData:     map[string]string{},
		},
		nil,
	)
	assertMapNestedStructBinding(t, registerSnapshot, "users", map[string]map[string]interface{}{
		"alice": {
			"Balance": "1",
			"Meta": map[string]interface{}{
				"Active": "true",
				"Limit":  "10",
			},
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
			Height:       base.Height(101),
			ContractCode: mapNestedStructEngineContractSource,
			Function:     "UpdateUser",
			CallData: map[string]string{
				"owner":   "bob",
				"balance": "9",
				"active":  "false",
				"limit":   "77",
			},
		},
		mustMarshalSnapshotDoc(t, registerSnapshot),
	)
	assertMapNestedStructBinding(t, callSnapshot, "users", map[string]map[string]interface{}{
		"alice": {
			"Balance": "1",
			"Meta": map[string]interface{}{
				"Active": "true",
				"Limit":  "10",
			},
		},
		"bob": {
			"Balance": "9",
			"Meta": map[string]interface{}{
				"Active": "false",
				"Limit":  "77",
			},
		},
	})
}

func assertNestedStructBinding(t *testing.T, doc SnapshotDoc, name string, expected map[string]interface{}) {
	t.Helper()

	binding := findSnapshotBinding(t, doc, name)
	assertStructSnapshotValue(t, binding.Value, expected)
}

func assertMapNestedStructBinding(t *testing.T, doc SnapshotDoc, name string, expected map[string]map[string]interface{}) {
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
		assertStructSnapshotValue(t, entry.Value, want)
	}
}

func assertStructSnapshotValue(t *testing.T, value SnapshotValue, expected map[string]interface{}) {
	t.Helper()

	if value.Kind != string(TypeStruct) {
		t.Fatalf("expected struct snapshot value, got %#v", value)
	}
	if len(value.Fields) != len(expected) {
		t.Fatalf("unexpected field count: %#v", value.Fields)
	}

	for _, field := range value.Fields {
		want, found := expected[field.Name]
		if !found {
			t.Fatalf("unexpected field %q", field.Name)
		}

		switch typedWant := want.(type) {
		case string:
			if field.Value.Kind != string(TypeScalar) || field.Value.Scalar != typedWant {
				t.Fatalf("unexpected scalar field %q: %#v", field.Name, field.Value)
			}
		case map[string]interface{}:
			switch field.Value.Kind {
			case string(TypeStruct):
				assertStructSnapshotValue(t, field.Value, typedWant)
			case string(TypeMap):
				assertMapSnapshotValue(t, field.Value, typedWant)
			default:
				t.Fatalf("unexpected composite field %q: %#v", field.Name, field.Value)
			}
		case []interface{}:
			assertSliceSnapshotValue(t, field.Value, typedWant)
		default:
			t.Fatalf("unsupported expected field type %T", want)
		}
	}
}

func assertMapSnapshotValue(t *testing.T, value SnapshotValue, expected map[string]interface{}) {
	t.Helper()

	if value.Kind != string(TypeMap) {
		t.Fatalf("expected map snapshot value, got %#v", value)
	}
	if value.IsNil {
		t.Fatalf("expected non-nil map snapshot")
	}
	if len(value.Entries) != len(expected) {
		t.Fatalf("unexpected entry count: %#v", value.Entries)
	}

	for _, entry := range value.Entries {
		want, found := expected[entry.Key]
		if !found {
			t.Fatalf("unexpected entry key %q", entry.Key)
		}

		switch typedWant := want.(type) {
		case string:
			if entry.Value.Kind != string(TypeScalar) || entry.Value.Scalar != typedWant {
				t.Fatalf("unexpected scalar entry %q: %#v", entry.Key, entry.Value)
			}
		case map[string]interface{}:
			assertStructSnapshotValue(t, entry.Value, typedWant)
		case []interface{}:
			assertSliceSnapshotValue(t, entry.Value, typedWant)
		default:
			t.Fatalf("unsupported expected map entry type %T", want)
		}
	}
}

func assertSliceSnapshotValue(t *testing.T, value SnapshotValue, expected []interface{}) {
	t.Helper()

	if value.Kind != string(TypeSlice) {
		t.Fatalf("expected slice snapshot value, got %#v", value)
	}
	if value.IsNil {
		t.Fatalf("expected non-nil slice snapshot")
	}
	if len(value.Items) != len(expected) {
		t.Fatalf("unexpected slice length: %#v", value.Items)
	}

	for i, item := range value.Items {
		want := expected[i]
		switch typedWant := want.(type) {
		case string:
			if item.Kind != string(TypeScalar) || item.Scalar != typedWant {
				t.Fatalf("unexpected scalar item %d: %#v", i, item)
			}
		case map[string]interface{}:
			assertStructSnapshotValue(t, item, typedWant)
		case []interface{}:
			assertSliceSnapshotValue(t, item, typedWant)
		default:
			t.Fatalf("unsupported expected slice item type %T", want)
		}
	}
}
