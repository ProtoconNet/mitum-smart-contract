package runtime

import (
	"fmt"
	"strings"
	"testing"
)

func TestAnalyzeContractSchemaComplexityAllowsNormalCompositeContract(t *testing.T) {
	source := `package contract

import "mitum/chain"

type Meta struct {
	Limit int64
	Flags map[string]bool
	Aliases []string
}

type User struct {
	Name string
	Meta Meta
}

var users map[string]User

func Initialize(ctx chain.WriteContext) error {
	return nil
}

func GetUsers(ctx chain.QueryContext) map[string]User {
	return users
}
`

	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
}

func TestAnalyzeContractSchemaComplexityBoundariesAllowed(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "import count",
			source: schemaComplexityImportCountSource(MaxContractSchemaImports),
		},
		{
			name:   "function count",
			source: schemaComplexityFunctionCountSource(MaxContractSchemaFunctions),
		},
		{
			name:   "persistent global count",
			source: schemaComplexityPersistentGlobalCountSource(MaxContractSchemaPersistentGlobals),
		},
		{
			name:   "struct count",
			source: schemaComplexityStructCountSource(MaxContractSchemaStructs),
		},
		{
			name:   "struct field count",
			source: schemaComplexityStructFieldCountSource(MaxContractSchemaStructFields),
		},
		{
			name:   "type nesting depth",
			source: schemaComplexityNestedStructSource(MaxContractSchemaTypeNestingDepth),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := AnalyzeContractSchema(tt.source); err != nil {
				t.Fatalf("AnalyzeContractSchema returned error at boundary: %v", err)
			}
		})
	}
}

func TestAnalyzeContractSchemaComplexityLimitsRejected(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantError string
	}{
		{
			name:      "import count",
			source:    schemaComplexityImportCountSource(MaxContractSchemaImports + 1),
			wantError: "max import count",
		},
		{
			name:      "function count",
			source:    schemaComplexityFunctionCountSource(MaxContractSchemaFunctions + 1),
			wantError: "max function count",
		},
		{
			name:      "persistent global count",
			source:    schemaComplexityPersistentGlobalCountSource(MaxContractSchemaPersistentGlobals + 1),
			wantError: "max persistent global count",
		},
		{
			name:      "struct count",
			source:    schemaComplexityStructCountSource(MaxContractSchemaStructs + 1),
			wantError: "max struct count",
		},
		{
			name:      "struct field count",
			source:    schemaComplexityStructFieldCountSource(MaxContractSchemaStructFields + 1),
			wantError: "max struct field count",
		},
		{
			name:      "type nesting depth",
			source:    schemaComplexityNestedStructSource(MaxContractSchemaTypeNestingDepth + 1),
			wantError: "max type nesting depth",
		},
		{
			name:      "total node count",
			source:    schemaComplexityTotalNodeCountExceededSource(),
			wantError: "max total node count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AnalyzeContractSchema(tt.source)
			if err == nil {
				t.Fatalf("expected AnalyzeContractSchema to reject %s limit", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error containing %q, got: %v", tt.wantError, err)
			}
		})
	}
}

func schemaComplexityImportCountSource(count int) string {
	var b strings.Builder
	b.WriteString("package contract\n\n")
	b.WriteString("import (\n")
	b.WriteString(`"mitum/chain"` + "\n")
	for i := 1; i < count; i++ {
		fmt.Fprintf(&b, "s%d \"strings\"\n", i)
	}
	b.WriteString(")\n\n")
	b.WriteString(schemaComplexityInitializeOnly())

	return b.String()
}

func schemaComplexityFunctionCountSource(count int) string {
	var b strings.Builder
	b.WriteString(schemaComplexityHeader())
	b.WriteString(schemaComplexityInitializeOnly())
	for i := 1; i < count; i++ {
		fmt.Fprintf(&b, "func Query%03d(ctx chain.QueryContext) string { return \"\" }\n", i)
	}

	return b.String()
}

func schemaComplexityPersistentGlobalCountSource(count int) string {
	var b strings.Builder
	b.WriteString(schemaComplexityHeader())
	b.WriteString("var (\n")
	for i := 0; i < count; i++ {
		fmt.Fprintf(&b, "g%03d string\n", i)
	}
	b.WriteString(")\n\n")
	b.WriteString(schemaComplexityInitializeOnly())

	return b.String()
}

func schemaComplexityStructCountSource(count int) string {
	var b strings.Builder
	b.WriteString(schemaComplexityHeader())
	for i := 0; i < count; i++ {
		fmt.Fprintf(&b, "type S%03d struct { Value string }\n", i)
	}
	b.WriteString("\n")
	b.WriteString(schemaComplexityInitializeOnly())

	return b.String()
}

func schemaComplexityStructFieldCountSource(count int) string {
	var b strings.Builder
	b.WriteString(schemaComplexityHeader())
	b.WriteString("type Big struct {\n")
	for i := 0; i < count; i++ {
		fmt.Fprintf(&b, "F%03d string\n", i)
	}
	b.WriteString("}\n\n")
	b.WriteString("var big Big\n\n")
	b.WriteString(schemaComplexityInitializeOnly())

	return b.String()
}

func schemaComplexityNestedStructSource(depth int) string {
	var b strings.Builder
	b.WriteString(schemaComplexityHeader())
	for i := 0; i < depth; i++ {
		if i == depth-1 {
			fmt.Fprintf(&b, "type S%03d struct { Value string }\n", i)
			continue
		}
		fmt.Fprintf(&b, "type S%03d struct { Next S%03d }\n", i, i+1)
	}
	b.WriteString("\nvar root S000\n\n")
	b.WriteString(schemaComplexityInitializeOnly())

	return b.String()
}

func schemaComplexityTotalNodeCountExceededSource() string {
	var b strings.Builder
	b.WriteString(schemaComplexityHeader())
	b.WriteString("func Initialize(ctx chain.WriteContext")
	for i := 0; i < MaxContractSchemaNodes; i++ {
		fmt.Fprintf(&b, ", p%04d string", i)
	}
	b.WriteString(") error { return nil }\n")

	return b.String()
}

func schemaComplexityHeader() string {
	return "package contract\n\nimport \"mitum/chain\"\n\n"
}

func schemaComplexityInitializeOnly() string {
	return "func Initialize(ctx chain.WriteContext) error { return nil }\n"
}
