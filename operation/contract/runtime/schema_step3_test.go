package runtime

import "testing"

func TestAnalyzeContractSchemaNestedStructGlobalUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Limits struct {
	Daily int64
}

type Config struct {
	Limits Limits
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected nested struct unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "config", "Limits", "flat struct globals require scalar fields only") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaStructFieldMapUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Users map[string]int64
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map field unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "config", "Users", "flat struct globals require scalar fields only") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaStructFieldSliceUnsupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Flags []string
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected slice field unsupported error")
	}

	if got := err.Error(); got == "" || !containsAll(got, "config", "Flags", "flat struct globals require scalar fields only") {
		t.Fatalf("unexpected error: %v", err)
	}
}
