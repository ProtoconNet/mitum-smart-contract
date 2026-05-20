package runtime

import (
	"strings"
	"testing"
)

const zeroArgWriteShapeContractSource = `package contract

import "mitum/chain"

func Initialize(ctx chain.ContractContext) error { return nil }
func Claim(ctx chain.ContractContext) error { return nil }
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

func Initialize(ctx chain.ContractContext) error { return nil }
func GetValue(ctx chain.ContractContext) string { return "ok" }
func GetMaybe(ctx chain.ContractContext) (string, bool) { return "ok", true }
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

func Initialize(ctx chain.ContractContext) error { return nil }
func Bad(ctx chain.ContractContext) (string, string) { return "", "" }
`)
	if err == nil {
		t.Fatal("expected non-bool second query result to be rejected")
	}
	if !strings.Contains(err.Error(), `function "Bad"`) ||
		!strings.Contains(err.Error(), "write(ctx, ...) error or query(ctx, ...) T[/bool]") {
		t.Fatalf("unexpected error for non-bool second result: %v", err)
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
