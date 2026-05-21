package runtime

import "testing"

func TestAnalyzeContractSchemaNestedNamedStructGlobalSupported(t *testing.T) {
	source := `package contract
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

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, owner string, daily int64, max int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaMapStringNestedStructGlobalSupported(t *testing.T) {
	source := `package contract
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

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, owner string, balance int64, active bool, limit int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaRecursiveStructUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Node struct {
	Next Node
}

var node Node

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value int64) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected recursive struct unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "node", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMutuallyRecursiveStructUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type A struct {
	B B
}

type B struct {
	A A
}

var root A

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value int64) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected mutually recursive struct unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "root", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaAnonymousNestedStructUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Meta struct {
		Count int64
	}
}

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value int64) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected anonymous nested struct unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "user", "Meta", "anonymous nested struct fields are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaNestedStructWithMapFieldSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Meta struct {
	Flags map[string]bool
}

type User struct {
	Meta Meta
}

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaNestedStructWithSliceFieldSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Meta struct {
	Tags []string
}

type User struct {
	Meta Meta
}

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value int64) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}
