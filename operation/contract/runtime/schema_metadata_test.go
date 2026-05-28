package runtime

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ProtoconNet/mitum-smart-contract/types"
	"github.com/ProtoconNet/mitum2/base"
)

func TestPersistedContractSchemaRoundTrip(t *testing.T) {
	resetContractSchemaCacheForTest()
	defer resetContractSchemaCacheForTest()

	schema, err := AnalyzeContractSchema(schemaReuseContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	persisted := NewPersistedContractSchema(schemaReuseContractSource, schema)
	if persisted.SchemaFormatVersion != types.CurrentSchemaFormatVersion {
		t.Fatalf("unexpected format version: %q", persisted.SchemaFormatVersion)
	}
	if persisted.SchemaRulesetVersion != CurrentSchemaRulesetVersion {
		t.Fatalf("unexpected ruleset version: %q", persisted.SchemaRulesetVersion)
	}
	if persisted.SourceHash != types.ContractSourceHash(schemaReuseContractSource) {
		t.Fatalf("unexpected source hash: %q", persisted.SourceHash)
	}

	roundTrip, ok := RuntimeSchemaFromPersisted(schemaReuseContractSource, &persisted)
	if !ok {
		t.Fatal("expected persisted schema to be reusable")
	}
	if !reflect.DeepEqual(roundTrip, schema) {
		t.Fatalf("schema roundtrip mismatch\ngot:  %#v\nwant: %#v", roundTrip, schema)
	}
}

func TestRuntimeSchemaFromPersistedRejectsMismatchesForFallback(t *testing.T) {
	schema, err := AnalyzeContractSchema(schemaReuseContractSource)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*types.PersistedContractSchema)
		source string
	}{
		{
			name: "format version mismatch",
			mutate: func(p *types.PersistedContractSchema) {
				p.SchemaFormatVersion = "contract-schema-format-v0"
			},
			source: schemaReuseContractSource,
		},
		{
			name: "ruleset version mismatch",
			mutate: func(p *types.PersistedContractSchema) {
				p.SchemaRulesetVersion = "typed-gno-ruleset-v0"
			},
			source: schemaReuseContractSource,
		},
		{
			name: "source hash mismatch",
			mutate: func(p *types.PersistedContractSchema) {
			},
			source: schemaReuseContractSource + "\n",
		},
		{
			name: "invalid schema payload",
			mutate: func(p *types.PersistedContractSchema) {
				p.Schema.Functions[0].Params[0].Type.Kind = "unknown"
			},
			source: schemaReuseContractSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetContractSchemaCacheForTest()
			defer resetContractSchemaCacheForTest()

			persisted := NewPersistedContractSchema(schemaReuseContractSource, schema)
			tt.mutate(&persisted)

			if got, ok := RuntimeSchemaFromPersisted(tt.source, &persisted); ok {
				t.Fatalf("expected persisted schema fallback, got reusable schema: %#v", got)
			}
		})
	}
}

func TestGnoEnginePersistedSchemaHitAvoidsAnalyzerOnColdCache(t *testing.T) {
	source := schemaReuseContractSource
	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
	persisted := NewPersistedContractSchema(source, schema)

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

	runtimeSchema, ok := RuntimeSchemaFromPersisted(source, &persisted)
	if !ok {
		t.Fatal("expected persisted schema to be reusable")
	}
	if count != 0 {
		t.Fatalf("RuntimeSchemaFromPersisted should not analyze source, got count %d", count)
	}

	engine := NewGnoEngine()
	contract := base.NewStringAddress("contractmeta0001")
	sender := base.NewStringAddress("sendermeta0001")
	states := map[string]base.State{}

	result, berr := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(states), ExecuteRequest{
		Mode:         InvocationModeRegister,
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(810),
		ContractCode: source,
		Schema:       &runtimeSchema,
		Function:     "Initialize",
		CallData:     map[string]string{},
	})
	if berr != nil {
		t.Fatalf("ExecuteContract returned error: %v", berr)
	}
	if count != 0 {
		t.Fatalf("ExecuteContract with persisted schema should not analyze source, got count %d", count)
	}

	applyStateMerges(states, base.Height(810), result.StateMerges)

	querySchema, ok := RuntimeSchemaFromPersisted(source, &persisted)
	if !ok {
		t.Fatal("expected persisted schema to be reusable for query")
	}
	qr, berr := engine.QueryContract(newRuntimeTestEncoders(t), stateGetter(states), QueryRequest{
		Contract:     contract,
		Sender:       sender,
		Height:       base.Height(811),
		ContractCode: source,
		Schema:       &querySchema,
		Function:     "GetValue",
		CallData:     map[string]string{},
	})
	if berr != nil {
		t.Fatalf("QueryContract returned error: %v", berr)
	}
	if count != 0 {
		t.Fatalf("QueryContract with persisted schema should not analyze source, got count %d", count)
	}
	if got, ok := qr.Result.(string); !ok || got != "hello" {
		t.Fatalf("unexpected query result: %#v", qr.Result)
	}
}

func TestGnoEnginePersistedSchemaMismatchFallsBackToAnalyzer(t *testing.T) {
	source := schemaReuseContractSource
	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*types.PersistedContractSchema)
		source string
	}{
		{
			name: "format version mismatch",
			mutate: func(p *types.PersistedContractSchema) {
				p.SchemaFormatVersion = "contract-schema-format-v0"
			},
			source: source,
		},
		{
			name: "ruleset version mismatch",
			mutate: func(p *types.PersistedContractSchema) {
				p.SchemaRulesetVersion = "typed-gno-ruleset-v0"
			},
			source: source,
		},
		{
			name: "source hash mismatch",
			mutate: func(p *types.PersistedContractSchema) {
			},
			source: source + "\n",
		},
		{
			name: "invalid schema payload",
			mutate: func(p *types.PersistedContractSchema) {
				p.Schema.Functions = nil
			},
			source: source,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			persisted := NewPersistedContractSchema(source, schema)
			tt.mutate(&persisted)
			persistedSchema, ok := RuntimeSchemaFromPersisted(tt.source, &persisted)
			if ok {
				t.Fatalf("expected persisted schema fallback, got %#v", persistedSchema)
			}

			engine := NewGnoEngine()
			contract := base.NewStringAddress(fmt.Sprintf("contractmetaf%04d", i))
			sender := base.NewStringAddress(fmt.Sprintf("sendermetaf%04d", i))
			_, berr := engine.ExecuteContract(newRuntimeTestEncoders(t), stateGetter(map[string]base.State{}), ExecuteRequest{
				Mode:         InvocationModeRegister,
				Contract:     contract,
				Sender:       sender,
				Height:       base.Height(820 + i),
				ContractCode: source,
				Function:     "Initialize",
				CallData:     map[string]string{},
			})
			if berr != nil {
				t.Fatalf("ExecuteContract fallback returned error: %v", berr)
			}
			if count != 1 {
				t.Fatalf("expected analyzer fallback once, got %d", count)
			}
		})
	}
}
