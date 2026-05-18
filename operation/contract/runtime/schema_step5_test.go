package runtime

import "testing"

func TestAnalyzeContractSchemaMapStringStructGlobalSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, owner string, amount int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}

	global := schema.PersistentGlobals[0]
	if global.Type.Kind != TypeMap {
		t.Fatalf("expected map type, got %#v", global.Type)
	}
	if global.Type.Elem == nil || global.Type.Elem.Kind != TypeNamed || global.Type.Elem.Name != "User" {
		t.Fatalf("expected map elem User, got %#v", global.Type.Elem)
	}
}

func TestAnalyzeContractSchemaMapStringStructWithMapFieldSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Flags map[string]bool
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, owner string, amount int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaMapStringStructWithSliceFieldSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Tags []string
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, owner string, amount int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}
