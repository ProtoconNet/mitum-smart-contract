package runtime

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestAnalyzeContractSchemaScalarStillSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

var initialized bool
var owner string
var revision int64

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
func GetRevision(ctx chain.ContractContext) int64 { return revision }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if schema.Mode != SchemaModeTypedArgs {
		t.Fatalf("unexpected schema mode: %s", schema.Mode)
	}

	if got := schema.PersistentGlobals[0].Type.Kind; got != TypeScalar {
		t.Fatalf("expected scalar persistent global, got %s", got)
	}
}

func TestTypeResolverRecognizesStructMapSlice(t *testing.T) {
	source := `package contract
type Config struct {
	Owner string
	Count int64
}
var cfg Config
var byOwner map[string]Config
var labels []string
`

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, parser.AllErrors)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	resolver := newTypeResolver(fset, node)

	varDecls := []*ast.GenDecl{}
	for _, decl := range node.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			varDecls = append(varDecls, gen)
		}
	}

	if len(varDecls) != 3 {
		t.Fatalf("expected 3 var decls, got %d", len(varDecls))
	}

	cfgBindings, err := parsePersistentBindings(resolver, varDecls[0])
	if err != nil {
		t.Fatalf("cfg binding parse: %v", err)
	}
	if cfgBindings[0].Type.Kind != TypeNamed || cfgBindings[0].Type.Name != "Config" {
		t.Fatalf("expected named Config type, got %#v", cfgBindings[0].Type)
	}

	mapBindings, err := parsePersistentBindings(resolver, varDecls[1])
	if err != nil {
		t.Fatalf("map binding parse: %v", err)
	}
	if mapBindings[0].Type.Kind != TypeMap {
		t.Fatalf("expected map type, got %#v", mapBindings[0].Type)
	}
	if mapBindings[0].Type.Key == nil || mapBindings[0].Type.Key.Kind != TypeScalar || mapBindings[0].Type.Key.NormalizedString() != "string" {
		t.Fatalf("expected string map key, got %#v", mapBindings[0].Type.Key)
	}
	if mapBindings[0].Type.Elem == nil || mapBindings[0].Type.Elem.Kind != TypeNamed || mapBindings[0].Type.Elem.Name != "Config" {
		t.Fatalf("expected Config map elem, got %#v", mapBindings[0].Type.Elem)
	}

	sliceBindings, err := parsePersistentBindings(resolver, varDecls[2])
	if err != nil {
		t.Fatalf("slice binding parse: %v", err)
	}
	if sliceBindings[0].Type.Kind != TypeSlice {
		t.Fatalf("expected slice type, got %#v", sliceBindings[0].Type)
	}
	if sliceBindings[0].Type.Elem == nil || sliceBindings[0].Type.Elem.Kind != TypeScalar || sliceBindings[0].Type.Elem.NormalizedString() != "string" {
		t.Fatalf("expected string slice elem, got %#v", sliceBindings[0].Type.Elem)
	}
}

func TestAnalyzeContractSchemaFlatStructGlobalSupported(t *testing.T) {
	source := `package contract
import "mitum/chain"

type Config struct {
	Owner string
	Paused bool
	Limit int64
}

var config Config

func Initialize(ctx chain.ContractContext) error { return nil }
func Update(ctx chain.ContractContext, value string) error { return nil }
`

	schema, err := AnalyzeContractSchema(source)
	if err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}

	if len(schema.PersistentGlobals) != 1 {
		t.Fatalf("expected 1 persistent global, got %d", len(schema.PersistentGlobals))
	}

	global := schema.PersistentGlobals[0]
	if global.Type.Kind != TypeNamed || global.Type.Name != "Config" {
		t.Fatalf("expected named Config global type, got %#v", global.Type)
	}

	resolved := schema.ResolveType(global.Type)
	if resolved.Kind != TypeStruct {
		t.Fatalf("expected resolved struct type, got %#v", resolved)
	}
	if len(resolved.Fields) != 3 {
		t.Fatalf("expected 3 struct fields, got %d", len(resolved.Fields))
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
