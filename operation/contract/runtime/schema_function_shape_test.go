package runtime

import (
	"strings"
	"testing"
)

const zeroArgWriteShapeContractSource = `package contract

import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func Claim(ctx chain.WriteContext) error { return nil }
`

func TestZeroArgWriteFunctionIsAcceptedAsWrite(t *testing.T) {
	schema, err := AnalyzeContractSchema(zeroArgWriteShapeContractSource)
	if err != nil {
		t.Fatalf("expected zero-arg write contract to analyze successfully, got: %v", err)
	}

	fn := mustFindFunctionSchema(t, schema, "Claim")
	if !fn.IsTypedWriteShape() {
		t.Fatal("expected Claim(ctx) error to be classified as typed write")
	}
	if fn.IsTypedQueryShape() {
		t.Fatal("expected Claim(ctx) error not to be classified as typed query")
	}
	if !fn.IsTypedABIShape() {
		t.Fatal("expected Claim(ctx) error to remain a typed ABI-shaped function")
	}
}

func TestQueryFunctionShapesStillAccepted(t *testing.T) {
	schema, err := AnalyzeContractSchema(`package contract

import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func GetValue(ctx chain.QueryContext) string { return "ok" }
func GetMaybe(ctx chain.QueryContext) (string, bool) { return "ok", true }
`)
	if err != nil {
		t.Fatalf("expected query shape contract to analyze successfully, got: %v", err)
	}

	single := mustFindFunctionSchema(t, schema, "GetValue")
	if !single.IsTypedQueryShape() {
		t.Fatal("expected single-result query to remain accepted")
	}
	if single.IsTypedWriteShape() {
		t.Fatal("expected single-result query not to be classified as write")
	}

	boolPair := mustFindFunctionSchema(t, schema, "GetMaybe")
	if !boolPair.IsTypedQueryShape() {
		t.Fatal("expected (T, bool) query to remain accepted")
	}
	if boolPair.IsTypedWriteShape() {
		t.Fatal("expected (T, bool) query not to be classified as write")
	}
}

func TestInvalidQuerySecondResultNonBoolStillRejected(t *testing.T) {
	_, err := AnalyzeContractSchema(`package contract

import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func Bad(ctx chain.QueryContext) (string, string) { return "", "" }
`)
	if err == nil {
		t.Fatal("expected non-bool second query result to be rejected")
	}
	if !strings.Contains(err.Error(), `query function "Bad" second result must be bool`) {
		t.Fatalf("unexpected error for non-bool second result: %v", err)
	}
}

func TestContextSplitRejectsMismatchedFunctionShapes(t *testing.T) {
	tests := []struct {
		name       string
		function   string
		wantErrSub string
	}{
		{
			name:       "query_result_with_write_context",
			function:   `func Get(ctx chain.WriteContext) string { return "nope" }`,
			wantErrSub: `write(ctx WriteContext, ...) error or query(ctx QueryContext, ...) T[/bool]`,
		},
		{
			name:       "write_result_with_query_context",
			function:   `func Claim(ctx chain.QueryContext) error { return nil }`,
			wantErrSub: `write(ctx WriteContext, ...) error or query(ctx QueryContext, ...) T[/bool]`,
		},
		{
			name:       "legacy_contract_context_write",
			function:   `func Claim(ctx chain.ContractContext) error { return nil }`,
			wantErrSub: `ContractContext is not supported`,
		},
		{
			name:       "legacy_contract_context_query",
			function:   `func Get(ctx chain.ContractContext) string { return "" }`,
			wantErrSub: `ContractContext is not supported`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AnalyzeContractSchema(`package contract

import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
` + tt.function + "\n")
			if err == nil {
				t.Fatal("expected context-shape mismatch to be rejected")
			}
			if !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInitializeRequiresWriteContext(t *testing.T) {
	_, err := AnalyzeContractSchema(`package contract

import "mitum/chain"

func Initialize(ctx chain.QueryContext) error { return nil }
`)
	if err == nil {
		t.Fatal("expected Initialize with QueryContext to be rejected")
	}
	if !strings.Contains(err.Error(), "Initialize(ctx WriteContext") {
		t.Fatalf("unexpected initialize context error: %v", err)
	}
}

func TestHostABIQueryContextHasNoSenderAccessor(t *testing.T) {
	if !strings.Contains(mitumChainPackageSource, "type WriteContext struct") {
		t.Fatal("host ABI package must define WriteContext")
	}
	if !strings.Contains(mitumChainPackageSource, "type QueryContext struct") {
		t.Fatal("host ABI package must define QueryContext")
	}
	if strings.Contains(mitumChainPackageSource, "func (ctx QueryContext) GetSender") {
		t.Fatal("QueryContext must not expose GetSender")
	}
}

func mustFindFunctionSchema(t *testing.T, schema ContractSchema, name string) FunctionSchema {
	t.Helper()

	fn, found := schema.FindFunction(name)
	if !found {
		t.Fatalf("function %q not found in schema", name)
	}

	return fn
}
