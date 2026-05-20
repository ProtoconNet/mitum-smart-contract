package runtime

import (
	"strings"
	"testing"
)

func TestGnoEngineExecuteRejectsOversizedCallDataBeforeSchemaAnalysis(t *testing.T) {
	resetContractSchemaCacheForTest()
	originalAnalyzer := analyzeContractSchemaFunc
	defer func() {
		analyzeContractSchemaFunc = originalAnalyzer
		resetContractSchemaCacheForTest()
	}()

	count := 0
	analyzeContractSchemaFunc = func(sourceCode string) (ContractSchema, error) {
		count++
		return AnalyzeContractSchema(sourceCode)
	}

	engine := NewGnoEngine()
	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), nil, ExecuteRequest{
		CallData: map[string]string{
			"value": strings.Repeat("v", MaxContractCallDataValueBytes+1),
		},
	})
	if err == nil {
		t.Fatal("expected ExecuteContract to reject oversized call data")
	}
	if !containsAll(err.Error(), "invalid call data", "value for key", "exceeds max size") {
		t.Fatalf("expected invalid call data error with core wording, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected schema analyzer not to run, got %d calls", count)
	}
}

func TestGnoEngineQueryRejectsOversizedCallDataBeforeSchemaAnalysis(t *testing.T) {
	resetContractSchemaCacheForTest()
	originalAnalyzer := analyzeContractSchemaFunc
	defer func() {
		analyzeContractSchemaFunc = originalAnalyzer
		resetContractSchemaCacheForTest()
	}()

	count := 0
	analyzeContractSchemaFunc = func(sourceCode string) (ContractSchema, error) {
		count++
		return AnalyzeContractSchema(sourceCode)
	}

	engine := NewGnoEngine()
	_, err := engine.QueryContract(newRuntimeTestEncoders(t), nil, QueryRequest{
		CallData: map[string]string{
			"value": strings.Repeat("v", MaxContractCallDataValueBytes+1),
		},
	})
	if err == nil {
		t.Fatal("expected QueryContract to reject oversized call data")
	}
	if !containsAll(err.Error(), "invalid call data", "value for key", "exceeds max size") {
		t.Fatalf("expected invalid call data error with core wording, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected schema analyzer not to run, got %d calls", count)
	}
}
