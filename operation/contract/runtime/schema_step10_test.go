package runtime

import "testing"

func TestAnalyzeContractSchemaStructArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Owner string
	Flags map[string]bool
	Aliases []string
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func SetConfig(ctx chain.ContractContext, cfg Config) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaMapStringScalarArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var flags map[string]bool

func Initialize(ctx chain.ContractContext) error { return nil }
func SetFlags(ctx chain.ContractContext, flags map[string]bool) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaMapStringNamedStructArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Aliases []string
}

var users map[string]User

func Initialize(ctx chain.ContractContext) error { return nil }
func ReplaceUsers(ctx chain.ContractContext, users map[string]User) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaSliceScalarArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var aliases []string

func Initialize(ctx chain.ContractContext) error { return nil }
func SetAliases(ctx chain.ContractContext, aliases []string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaSliceNamedStructArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type User struct {
	Balance int64
	Active bool
}

var watchers []User

func Initialize(ctx chain.ContractContext) error { return nil }
func ReplaceWatchers(ctx chain.ContractContext, watchers []User) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	if len(schema.Functions) == 0 {
		t.Fatalf("expected parsed functions")
	}
}

func TestAnalyzeContractSchemaAnonymousStructArgUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func SetInline(ctx chain.ContractContext, values []struct { Count int64 }) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected anonymous struct arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "SetInline", "anonymous struct slice elements are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaSliceMapArgUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func SetMatrix(ctx chain.ContractContext, matrix []map[string]int64) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected []map[string]int64 arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "SetMatrix", "slice elements cannot be maps") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMapSliceArgUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func SetNames(ctx chain.ContractContext, names map[string][]string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map[string][]string arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "SetNames", "map values cannot be slices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRecursiveArgTypeUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Node struct {
	Children []Node
}

var root Node

func Initialize(ctx chain.ContractContext) error { return nil }
func SetRoot(ctx chain.ContractContext, root Node) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected recursive arg type unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "root", "recursive named struct types are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
