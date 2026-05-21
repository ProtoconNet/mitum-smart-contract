package runtime

import "testing"

func TestAnalyzeContractSchemaTopLevelSliceScalarSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var flags []bool

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaTopLevelSliceNamedStructSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users []User

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaStructFieldSliceNamedStructSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Meta struct {
	Limit int64
}

type User struct {
	Meta Meta
}

type Config struct {
	Users []User
	Names []string
}

var config Config

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}
}

func TestAnalyzeContractSchemaTopLevelNestedSliceUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var groups [][]string

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected [][]string unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "groups", "slice elements cannot be slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaTopLevelSliceMapElemUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var matrix []map[string]int64

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected []map[string]T unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "matrix", "slice elements cannot be maps") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaAnonymousStructSliceUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Inline []struct {
		Count int64
	}
}

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected []anonymous-struct unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "user", "anonymous struct slice elements are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRecursiveStructViaSliceUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Node struct {
	Children []Node
}

var root Node

func Initialize(ctx chain.WriteContext) error { return nil }
func Update(ctx chain.WriteContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected recursive struct via slice unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "root", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaSliceQueryResultSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var names []string

func Initialize(ctx chain.WriteContext) error { return nil }
func GetNames(ctx chain.QueryContext) []string { return names }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}
