package runtime

import (
	"strings"
	"testing"

	"github.com/ProtoconNet/mitum2/base"
)

const schemaReuseContractSource = `package contract

import "mitum/chain"

var value string

func Initialize(ctx chain.ContractContext) error {
	value = "hello"
	return nil
}

func GetValue(ctx chain.ContractContext) string {
	return value
}
`

const schemaReuseDifferentContractSource = `package contract

import "mitum/chain"

var value string

func Initialize(ctx chain.ContractContext) error {
	value = "different"
	return nil
}

func GetValue(ctx chain.ContractContext) string {
	return value
}
`

func TestGnoEngineRepeatedValidationReusesSchemaCache(t *testing.T) {
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
	if _, err := engine.ValidateContract(schemaReuseContractSource); err != nil {
		t.Fatalf("first ValidateContract returned error: %v", err)
	}
	if _, err := engine.ValidateContract(schemaReuseContractSource); err != nil {
		t.Fatalf("second ValidateContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected repeated ValidateContract to analyze once, got %d", count)
	}
}

func TestGnoEngineValidationThenExecuteReusesSchemaCache(t *testing.T) {
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
	if _, err := engine.ValidateContract(schemaReuseContractSource); err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected ValidateContract to analyze once, got %d", count)
	}

	contract := base.NewStringAddress("contractreuse0002")
	sender := base.NewStringAddress("senderreuse0002")
	states := map[string]base.State{}

	_, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(710),
		ContractCode: schemaReuseContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected ExecuteContract to reuse cached schema, got analyze count %d", count)
	}
}

func TestGnoEngineValidationThenQueryReusesSchemaCache(t *testing.T) {
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
	if _, err := engine.ValidateContract(schemaReuseContractSource); err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected ValidateContract to analyze once, got %d", count)
	}

	contract := base.NewStringAddress("contractreuse0003")
	sender := base.NewStringAddress("senderreuse0003")
	states := map[string]base.State{}

	result, err := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(720),
		ContractCode: schemaReuseContractSource,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("ExecuteContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected ExecuteContract to reuse cached schema, got analyze count %d", count)
	}

	applyStateMerges(states, base.Height(720), result.StateMerges)

	qr, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(721),
		ContractCode: schemaReuseContractSource,
		Function:     "GetValue",
		CallData:     map[string]string{},
	})
	if err != nil {
		t.Fatalf("QueryContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected QueryContract to reuse cached schema, got analyze count %d", count)
	}
	if got, ok := qr.Result.(string); !ok || got != "hello" {
		t.Fatalf("unexpected query result: %#v", qr.Result)
	}
}

func TestGnoEngineDifferentSourceUsesDifferentSchemaCacheEntry(t *testing.T) {
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
	if _, err := engine.ValidateContract(schemaReuseContractSource); err != nil {
		t.Fatalf("ValidateContract(source A) returned error: %v", err)
	}
	if _, err := engine.ValidateContract(schemaReuseDifferentContractSource); err != nil {
		t.Fatalf("ValidateContract(source B) returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected different sources to analyze separately, got %d", count)
	}
}

func TestGnoEngineRequestSchemaTakesPrecedenceOverSchemaCache(t *testing.T) {
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
	if _, err := engine.ValidateContract(schemaReuseContractSource); err != nil {
		t.Fatalf("ValidateContract returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected ValidateContract to analyze once, got %d", count)
	}

	provided := ContractSchema{
		PackageName: "contract",
		Mode:        SchemaModeTypedArgs,
		Types:       NewTypeRegistry(),
	}
	_, err := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), QueryRequest{
		Contract:     base.NewStringAddress("contractreuse0004"),
		Sender:       base.NewStringAddress("senderreuse0004"),
		Height:       base.Height(730),
		ContractCode: schemaReuseContractSource,
		Schema:       &provided,
		Function:     "GetValue",
		CallData:     map[string]string{},
	})
	if err == nil {
		t.Fatal("expected QueryContract to fail with the provided schema")
	}
	if !strings.Contains(err.Error(), `query function "GetValue" not found`) {
		t.Fatalf("expected provided schema to take precedence, got error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected provided schema to avoid cache/analyzer lookup, got analyze count %d", count)
	}
}
