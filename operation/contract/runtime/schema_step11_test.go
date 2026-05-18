package runtime

import "testing"

func TestAnalyzeContractSchemaStructQueryArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Flags map[string]bool
	Aliases []string
}

type Selector struct {
	Name string
	WantActive bool
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func GetSelectedUser(ctx chain.ContractContext, selector Selector) (User, bool) { return User{}, false }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaMapStringScalarQueryArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoFlags(ctx chain.ContractContext, flags map[string]bool) map[string]bool { return flags }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaMapStringNamedStructQueryArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Aliases []string
}

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoUsers(ctx chain.ContractContext, users map[string]User) map[string]User { return users }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaSliceScalarQueryArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoAliases(ctx chain.ContractContext, aliases []string) []string { return aliases }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaSliceNamedStructQueryArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoWatchers(ctx chain.ContractContext, watchers []User) []User { return watchers }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaAnonymousStructQueryArgUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoInline(ctx chain.ContractContext, values []struct { Count int64 }) []string { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected anonymous struct query arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "EchoInline", "anonymous struct slice elements are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaSliceMapQueryArgUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoMatrix(ctx chain.ContractContext, matrix []map[string]int64) bool { return true }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected []map[string]int64 query arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "EchoMatrix", "slice elements cannot be maps") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMapSliceQueryArgUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoNames(ctx chain.ContractContext, names map[string][]string) bool { return true }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map[string][]string query arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "EchoNames", "map values cannot be slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRecursiveQueryArgTypeUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Node struct {
	Children []Node
}

func Initialize(ctx chain.ContractContext) error { return nil }
func GetNode(ctx chain.ContractContext, node Node) bool { return true }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected recursive query arg type unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "GetNode", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
