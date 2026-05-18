package runtime

import "testing"

func TestAnalyzeContractSchemaScalarQueryArgSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func HasAlias(ctx chain.ContractContext, name string) bool { return true }
`

	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
}

func TestAnalyzeContractSchemaStructQueryArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Selector struct { Name string }
type User struct { Balance int64 }

func Initialize(ctx chain.ContractContext) error { return nil }
func GetUser(ctx chain.ContractContext, selector Selector) (User, bool) { return User{}, false }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected struct query arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "GetUser", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaMapQueryArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoFlags(ctx chain.ContractContext, flags map[string]bool) map[string]bool { return flags }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected map query arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "EchoFlags", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaSliceQueryArgRejected(t *testing.T) {
	source := `package contract
import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func EchoAliases(ctx chain.ContractContext, aliases []string) []string { return aliases }
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected slice query arg unsupported error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "EchoAliases", "supports only scalar parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}
