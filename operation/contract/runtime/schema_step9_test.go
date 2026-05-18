package runtime

import "testing"

func TestAnalyzeContractSchemaMapStringNamedStructQueryResultSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func GetUsers(ctx chain.ContractContext) map[string]User { return users }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaSliceNamedStructQueryResultSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var watchers []User

func Initialize(ctx chain.ContractContext) error { return nil }
func GetWatchers(ctx chain.ContractContext) []User { return watchers }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaAnonymousStructQueryResultUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func GetInline(ctx chain.ContractContext) struct { Count int64 } { return struct { Count int64 }{Count:1} }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected anonymous struct query result unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "GetInline", "query result", "not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaSliceMapQueryResultUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func GetMatrix(ctx chain.ContractContext) []map[string]int64 { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected []map[string]int64 query result unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "GetMatrix", "slice elements cannot be maps") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMapSliceQueryResultUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func GetNames(ctx chain.ContractContext) map[string][]string { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map[string][]string query result unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "GetNames", "map values cannot be slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRecursiveQueryResultUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Node struct {
	Children []Node
}

var root Node

func Initialize(ctx chain.ContractContext) error { return nil }
func GetRoot(ctx chain.ContractContext) Node { return root }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected recursive query result unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "root", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
