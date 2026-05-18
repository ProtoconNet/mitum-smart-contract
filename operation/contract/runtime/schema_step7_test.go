package runtime

import "testing"

func TestAnalyzeContractSchemaNestedStructFieldMapStringScalarSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Flags map[string]bool
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaStructFieldMapStringNamedStructSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Meta struct {
	Limit int64
}

type User struct {
	Meta Meta
}

type Config struct {
	Users map[string]User
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaStructFieldMapNonStringKeyUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Extra map[int]string
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected non-string map key unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "config", "int", "map key type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaStructFieldMapStringSliceElemUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Extra map[string][]string
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map[string][]T unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "config", "[]string", "map values cannot be slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRecursiveStructViaMapFieldUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Node struct {
	Children map[string]Node
}

var root Node

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected recursive struct via map field unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "root", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
