package runtime

import (
	"reflect"
	"strings"
	"testing"
)

func TestCurrentSchemaRulesetMatchesCurrentTypedGnoPolicy(t *testing.T) {
	rules := CurrentSchemaRuleset()

	if rules.Version != CurrentSchemaRulesetVersion {
		t.Fatalf("unexpected ruleset version: %q", rules.Version)
	}

	if rules.SourceRules.RequiredPackageName != "contract" {
		t.Fatalf("unexpected required package name: %q", rules.SourceRules.RequiredPackageName)
	}
	if rules.SourceRules.MaxContractSourceBytes != MaxTypedContractSourceBytes {
		t.Fatalf("unexpected max contract bytes: %d", rules.SourceRules.MaxContractSourceBytes)
	}

	expectedImports := []string{
		MitumChainPackagePath,
		"strconv",
		"strings",
		"errors",
		"bytes",
		"encoding/hex",
		"encoding/base64",
		"unicode/utf8",
	}
	if !reflect.DeepEqual(rules.ImportRules.AllowedImports, expectedImports) {
		t.Fatalf("unexpected allowed imports: %#v", rules.ImportRules.AllowedImports)
	}

	expectedScalarKinds := []string{"string", "bool", "int", "int64", "uint64"}
	if !reflect.DeepEqual(rules.ScalarRules.AllowedKinds, expectedScalarKinds) {
		t.Fatalf("unexpected scalar kinds: %#v", rules.ScalarRules.AllowedKinds)
	}
	if rules.ContextRules.WriteContextType != "WriteContext" ||
		rules.ContextRules.QueryContextType != "QueryContext" ||
		rules.ContextRules.LegacyContractContextType != "ContractContext" ||
		rules.ContextRules.LegacyContractContextAllowed ||
		rules.ContextRules.QueryContextSenderAllowed ||
		rules.ContextRules.QuerySenderCallDataKeyParsed ||
		!rules.ContextRules.QueryContextCurrentHeightAllowed ||
		rules.ContextRules.WriteContextCurrentHeightAllowed ||
		rules.ContextRules.ChainCurrentHeightNativeAllowed {
		t.Fatalf("unexpected context rules: %#v", rules.ContextRules)
	}
	if rules.InputRules.CompositeInputAllowed {
		t.Fatal("expected composite input to remain disabled")
	}

	expectedTopLevelKinds := []SchemaShapeKind{
		SchemaShapeScalar,
		SchemaShapeNamedStruct,
		SchemaShapeMap,
		SchemaShapeSlice,
	}
	expectedElementKinds := []SchemaShapeKind{
		SchemaShapeScalar,
		SchemaShapeNamedStruct,
	}
	if !reflect.DeepEqual(rules.StateRules.TopLevelPersistentGlobalKinds, expectedTopLevelKinds) {
		t.Fatalf("unexpected top-level state kinds: %#v", rules.StateRules.TopLevelPersistentGlobalKinds)
	}
	if !reflect.DeepEqual(rules.StateRules.NestedStructFieldKinds, expectedTopLevelKinds) {
		t.Fatalf("unexpected nested state field kinds: %#v", rules.StateRules.NestedStructFieldKinds)
	}
	if !reflect.DeepEqual(rules.StateRules.SliceElementKinds, expectedElementKinds) {
		t.Fatalf("unexpected state slice element kinds: %#v", rules.StateRules.SliceElementKinds)
	}
	if !reflect.DeepEqual(rules.StateRules.MapKeyScalarKinds, []string{"string"}) {
		t.Fatalf("unexpected state map key kinds: %#v", rules.StateRules.MapKeyScalarKinds)
	}
	if !reflect.DeepEqual(rules.StateRules.MapValueKinds, expectedElementKinds) {
		t.Fatalf("unexpected state map value kinds: %#v", rules.StateRules.MapValueKinds)
	}
	if !rules.StateRules.NamedStructAllowed ||
		!rules.StateRules.NestedNamedStructAllowed ||
		!rules.StateRules.ExplicitPersistentGlobalTypeRequired ||
		rules.StateRules.EmbeddedFieldsAllowed ||
		rules.StateRules.AnonymousNestedStructFieldsAllowed ||
		rules.StateRules.RecursiveNamedStructsAllowed ||
		!rules.StateRules.TopLevelMapAllowed ||
		!rules.StateRules.NestedMapAllowed ||
		!rules.StateRules.MapStringScalarAllowed ||
		!rules.StateRules.MapStringNamedStructAllowed ||
		rules.StateRules.AnonymousStructMapValuesAllowed ||
		rules.StateRules.MapValueSlicesAllowed ||
		rules.StateRules.MapValueMapsAllowed ||
		!rules.StateRules.TopLevelSliceAllowed ||
		!rules.StateRules.NestedSliceAllowed ||
		!rules.StateRules.SliceScalarAllowed ||
		!rules.StateRules.SliceNamedStructAllowed ||
		rules.StateRules.AnonymousStructSliceElementsAllowed ||
		rules.StateRules.SliceElementSlicesAllowed ||
		rules.StateRules.SliceElementMapsAllowed {
		t.Fatalf("unexpected state rules: %#v", rules.StateRules)
	}

	if !reflect.DeepEqual(rules.QueryRules.TopLevelResultKinds, expectedTopLevelKinds) {
		t.Fatalf("unexpected query result kinds: %#v", rules.QueryRules.TopLevelResultKinds)
	}
	if !reflect.DeepEqual(rules.QueryRules.BoolPresenceResultValueKinds, expectedTopLevelKinds) {
		t.Fatalf("unexpected query bool-pair result value kinds: %#v", rules.QueryRules.BoolPresenceResultValueKinds)
	}
	if !reflect.DeepEqual(rules.QueryRules.NestedStructFieldKinds, expectedTopLevelKinds) {
		t.Fatalf("unexpected nested query field kinds: %#v", rules.QueryRules.NestedStructFieldKinds)
	}
	if !reflect.DeepEqual(rules.QueryRules.SliceElementKinds, expectedElementKinds) {
		t.Fatalf("unexpected query slice element kinds: %#v", rules.QueryRules.SliceElementKinds)
	}
	if !reflect.DeepEqual(rules.QueryRules.MapKeyScalarKinds, []string{"string"}) {
		t.Fatalf("unexpected query map key kinds: %#v", rules.QueryRules.MapKeyScalarKinds)
	}
	if !reflect.DeepEqual(rules.QueryRules.MapValueKinds, expectedElementKinds) {
		t.Fatalf("unexpected query map value kinds: %#v", rules.QueryRules.MapValueKinds)
	}
	if !rules.QueryRules.NamedStructAllowed ||
		!rules.QueryRules.NestedNamedStructAllowed ||
		!rules.QueryRules.SingleResultAllowed ||
		!rules.QueryRules.BoolPresenceResultAllowed ||
		rules.QueryRules.BoolPresenceSecondResultScalarKind != "bool" ||
		rules.QueryRules.EmbeddedFieldsAllowed ||
		rules.QueryRules.AnonymousResultStructAllowed ||
		rules.QueryRules.AnonymousNestedStructFieldsAllowed ||
		rules.QueryRules.RecursiveNamedStructsAllowed ||
		!rules.QueryRules.TopLevelMapAllowed ||
		!rules.QueryRules.NestedMapAllowed ||
		!rules.QueryRules.MapStringScalarAllowed ||
		!rules.QueryRules.MapStringNamedStructAllowed ||
		rules.QueryRules.AnonymousStructMapValuesAllowed ||
		rules.QueryRules.MapValueSlicesAllowed ||
		rules.QueryRules.MapValueMapsAllowed ||
		!rules.QueryRules.TopLevelSliceAllowed ||
		!rules.QueryRules.NestedSliceAllowed ||
		!rules.QueryRules.SliceScalarAllowed ||
		!rules.QueryRules.SliceNamedStructAllowed ||
		rules.QueryRules.AnonymousStructSliceElementsAllowed ||
		rules.QueryRules.SliceElementSlicesAllowed ||
		rules.QueryRules.SliceElementMapsAllowed {
		t.Fatalf("unexpected query rules: %#v", rules.QueryRules)
	}

	if !rules.LifecycleRules.InitializeRegisterOnly {
		t.Fatal("expected Initialize to remain register-only")
	}
	if rules.LifecycleRules.InitializeCallable {
		t.Fatal("expected Initialize to remain non-callable")
	}

	if rules.ComplexityRules.MaxImports != MaxContractSchemaImports ||
		rules.ComplexityRules.MaxFunctions != MaxContractSchemaFunctions ||
		rules.ComplexityRules.MaxPersistentGlobals != MaxContractSchemaPersistentGlobals ||
		rules.ComplexityRules.MaxStructs != MaxContractSchemaStructs ||
		rules.ComplexityRules.MaxStructFields != MaxContractSchemaStructFields ||
		rules.ComplexityRules.MaxTypeNestingDepth != MaxContractSchemaTypeNestingDepth ||
		rules.ComplexityRules.MaxSchemaNodes != MaxContractSchemaNodes {
		t.Fatalf("unexpected complexity rules: %#v", rules.ComplexityRules)
	}
}

func TestCurrentSchemaRulesetReturnsDefensiveCopies(t *testing.T) {
	rules := CurrentSchemaRuleset()
	rules.ImportRules.AllowedImports[0] = "fmt"
	rules.ScalarRules.AllowedKinds[0] = "float64"
	rules.StateRules.TopLevelPersistentGlobalKinds[0] = SchemaShapeAnonymousStruct
	rules.StateRules.NestedStructFieldKinds[0] = SchemaShapeAnonymousStruct
	rules.StateRules.SliceElementKinds[0] = SchemaShapeSlice
	rules.StateRules.MapKeyScalarKinds[0] = "int"
	rules.StateRules.MapValueKinds[0] = SchemaShapeMap
	rules.QueryRules.TopLevelResultKinds[0] = SchemaShapeAnonymousStruct
	rules.QueryRules.BoolPresenceResultValueKinds[0] = SchemaShapeAnonymousStruct
	rules.QueryRules.NestedStructFieldKinds[0] = SchemaShapeAnonymousStruct
	rules.QueryRules.SliceElementKinds[0] = SchemaShapeSlice
	rules.QueryRules.MapKeyScalarKinds[0] = "int"
	rules.QueryRules.MapValueKinds[0] = SchemaShapeMap

	fresh := CurrentSchemaRuleset()

	if fresh.ImportRules.AllowedImports[0] != MitumChainPackagePath {
		t.Fatalf("expected allowed imports to remain canonical, got %#v", fresh.ImportRules.AllowedImports)
	}
	if fresh.ScalarRules.AllowedKinds[0] != "string" {
		t.Fatalf("expected scalar kinds to remain canonical, got %#v", fresh.ScalarRules.AllowedKinds)
	}
	if fresh.StateRules.TopLevelPersistentGlobalKinds[0] != SchemaShapeScalar ||
		fresh.StateRules.NestedStructFieldKinds[0] != SchemaShapeScalar ||
		fresh.StateRules.SliceElementKinds[0] != SchemaShapeScalar ||
		fresh.StateRules.MapKeyScalarKinds[0] != "string" ||
		fresh.StateRules.MapValueKinds[0] != SchemaShapeScalar {
		t.Fatalf("expected state shape slices to remain canonical, got %#v", fresh.StateRules)
	}
	if fresh.QueryRules.TopLevelResultKinds[0] != SchemaShapeScalar ||
		fresh.QueryRules.BoolPresenceResultValueKinds[0] != SchemaShapeScalar ||
		fresh.QueryRules.NestedStructFieldKinds[0] != SchemaShapeScalar ||
		fresh.QueryRules.SliceElementKinds[0] != SchemaShapeScalar ||
		fresh.QueryRules.MapKeyScalarKinds[0] != "string" ||
		fresh.QueryRules.MapValueKinds[0] != SchemaShapeScalar {
		t.Fatalf("expected query shape slices to remain canonical, got %#v", fresh.QueryRules)
	}
}

func TestSchemaRulesetScalarRulesAreCanonicalValidatorSource(t *testing.T) {
	original := currentSchemaRuleset
	t.Cleanup(func() {
		currentSchemaRuleset = original
	})

	currentSchemaRuleset = CurrentSchemaRuleset()
	currentSchemaRuleset.ScalarRules.AllowedKinds = []string{"string"}

	if !IsSupportedScalarType(TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"}) {
		t.Fatal("expected string to remain supported by modified scalar rules")
	}
	if IsSupportedScalarType(TypeRef{Kind: TypeScalar, Raw: "int64", Scalar: "int64"}) {
		t.Fatal("expected int64 support to follow ScalarRules.AllowedKinds")
	}
}

func TestSchemaRulesetContextRulesAreCanonicalValidatorSource(t *testing.T) {
	original := currentSchemaRuleset
	t.Cleanup(func() {
		currentSchemaRuleset = original
	})

	currentSchemaRuleset = CurrentSchemaRuleset()
	currentSchemaRuleset.ContextRules.WriteContextType = "MutatingContext"
	currentSchemaRuleset.ContextRules.QueryContextType = "ReadContext"
	currentSchemaRuleset.ContextRules.LegacyContractContextAllowed = false

	writeFn := FunctionSchema{
		Name:     "Claim",
		Exported: true,
		Params: []ParamSchema{
			{Name: "ctx", Type: TypeRef{Kind: TypeOpaque, Raw: "chain.MutatingContext"}},
		},
		Results: []ResultSchema{
			{Type: TypeRef{Kind: TypeOpaque, Raw: "error"}},
		},
	}
	if !writeFn.IsTypedWriteShape() {
		t.Fatal("expected write shape to follow ContextRules.WriteContextType")
	}

	queryFn := FunctionSchema{
		Name:     "Get",
		Exported: true,
		Params: []ParamSchema{
			{Name: "ctx", Type: TypeRef{Kind: TypeOpaque, Raw: "chain.ReadContext"}},
		},
		Results: []ResultSchema{
			{Type: TypeRef{Kind: TypeScalar, Raw: "string", Scalar: "string"}},
		},
	}
	if !queryFn.IsTypedQueryShape() {
		t.Fatal("expected query shape to follow ContextRules.QueryContextType")
	}

	legacyFn := FunctionSchema{
		Name:     "Old",
		Exported: true,
		Params: []ParamSchema{
			{Name: "ctx", Type: TypeRef{Kind: TypeOpaque, Raw: "chain.ContractContext"}},
		},
	}
	if !legacyFn.UsesLegacyContractContext() {
		t.Fatal("expected legacy context detection to follow ContextRules")
	}
}

func TestSchemaRulesetCurrentHeightContextRulesAreValidatorSource(t *testing.T) {
	original := currentSchemaRuleset
	t.Cleanup(func() {
		currentSchemaRuleset = original
	})

	currentSchemaRuleset = CurrentSchemaRuleset()
	currentSchemaRuleset.ContextRules.QueryContextCurrentHeightAllowed = false

	_, err := AnalyzeContractSchema(`package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func GetCurrentHeight(ctx chain.QueryContext) int64 { return ctx.GetCurrentHeight() }
`)
	if err == nil || !strings.Contains(err.Error(), "QueryContext.GetCurrentHeight") {
		t.Fatalf("expected query current height rule rejection, got %v", err)
	}
}

func TestCurrentSchemaRulesetShapeFlagsMatchShapeKindSets(t *testing.T) {
	rules := CurrentSchemaRuleset()

	assertStateShapeFlagConsistency(t, rules.StateRules)
	assertQueryShapeFlagConsistency(t, rules.QueryRules)
}

func TestCurrentSchemaRulesetCompletenessMatchesUnsupportedStateShapes(t *testing.T) {
	rules := CurrentSchemaRuleset()

	if rules.StateRules.EmbeddedFieldsAllowed {
		t.Fatal("test expects embedded fields to be disallowed by current ruleset")
	}
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type Meta struct { Count int64 }
type User struct { Meta }

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
`, "embedded struct fields are not supported")

	if rules.StateRules.AnonymousNestedStructFieldsAllowed {
		t.Fatal("test expects anonymous nested struct fields to be disallowed by current ruleset")
	}
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type User struct {
	Meta struct { Count int64 }
}

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
`, "anonymous nested struct fields are not supported")

	if rules.StateRules.RecursiveNamedStructsAllowed {
		t.Fatal("test expects recursive named structs to be disallowed by current ruleset")
	}
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type Node struct { Children []Node }

var root Node

func Initialize(ctx chain.WriteContext) error { return nil }
`, "recursive named struct types are not supported")

	if rules.StateRules.AnonymousStructSliceElementsAllowed ||
		rules.StateRules.SliceElementSlicesAllowed ||
		rules.StateRules.SliceElementMapsAllowed {
		t.Fatal("test expects current slice element restrictions to be disallowed by ruleset")
	}
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type User struct {
	Inline []struct { Count int64 }
}

var user User

func Initialize(ctx chain.WriteContext) error { return nil }
`, "anonymous struct slice elements are not supported")
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

var groups [][]string

func Initialize(ctx chain.WriteContext) error { return nil }
`, "slice elements cannot be slices")
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

var matrix []map[string]int64

func Initialize(ctx chain.WriteContext) error { return nil }
`, "slice elements cannot be maps")

	if rules.StateRules.AnonymousStructMapValuesAllowed ||
		rules.StateRules.MapValueSlicesAllowed ||
		rules.StateRules.MapValueMapsAllowed {
		t.Fatal("test expects current map value restrictions to be disallowed by ruleset")
	}
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type Config struct {
	Inline map[string]struct { Count int64 }
}

var config Config

func Initialize(ctx chain.WriteContext) error { return nil }
`, "anonymous struct map values are not supported")
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type Config struct {
	Extra map[string][]string
}

var config Config

func Initialize(ctx chain.WriteContext) error { return nil }
`, "map values cannot be slices")
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

type Config struct {
	Extra map[string]map[string]string
}

var config Config

func Initialize(ctx chain.WriteContext) error { return nil }
`, "map values cannot be maps")
}

func TestCurrentSchemaRulesetCompletenessMatchesQueryResultPolicy(t *testing.T) {
	rules := CurrentSchemaRuleset()

	if !rules.QueryRules.SingleResultAllowed ||
		!rules.QueryRules.BoolPresenceResultAllowed ||
		rules.QueryRules.BoolPresenceSecondResultScalarKind != "bool" {
		t.Fatalf("unexpected query result form rules: %#v", rules.QueryRules)
	}

	if _, err := AnalyzeContractSchema(`package contract
import "mitum/chain"

type User struct { Balance int64 }
var users map[string]User

func Initialize(ctx chain.WriteContext) error { return nil }
func GetUsers(ctx chain.QueryContext) (map[string]User, bool) { return users, true }
`); err != nil {
		t.Fatalf("expected (T, bool) query result to be supported by current ruleset, got: %v", err)
	}

	if rules.QueryRules.AnonymousResultStructAllowed {
		t.Fatal("test expects anonymous query result structs to be disallowed by current ruleset")
	}
	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func GetInline(ctx chain.QueryContext) (struct { Count int64 }, bool) {
	return struct { Count int64 }{Count: 1}, true
}
`, "query result", "not supported")

	assertAnalyzeContractSchemaRejected(t, `package contract
import "mitum/chain"

func Initialize(ctx chain.WriteContext) error { return nil }
func GetValue(ctx chain.QueryContext) (string, string) { return "", "" }
`, `query function "GetValue" second result must be bool`)
}

func assertAnalyzeContractSchemaRejected(t *testing.T, source string, parts ...string) {
	t.Helper()

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected AnalyzeContractSchema to reject source")
	}
	if !containsAll(err.Error(), parts...) {
		t.Fatalf("expected error containing %q, got: %v", parts, err)
	}
}

func assertStateShapeFlagConsistency(t *testing.T, rules StateRules) {
	t.Helper()

	if rules.NamedStructAllowed != containsSchemaShapeKind(rules.TopLevelPersistentGlobalKinds, SchemaShapeNamedStruct) {
		t.Fatalf("state named-struct flag disagrees with top-level kinds: %#v", rules)
	}
	if rules.NestedNamedStructAllowed != containsSchemaShapeKind(rules.NestedStructFieldKinds, SchemaShapeNamedStruct) {
		t.Fatalf("state nested named-struct flag disagrees with nested field kinds: %#v", rules)
	}
	if rules.TopLevelMapAllowed != containsSchemaShapeKind(rules.TopLevelPersistentGlobalKinds, SchemaShapeMap) {
		t.Fatalf("state top-level map flag disagrees with top-level kinds: %#v", rules)
	}
	if rules.NestedMapAllowed != containsSchemaShapeKind(rules.NestedStructFieldKinds, SchemaShapeMap) {
		t.Fatalf("state nested map flag disagrees with nested field kinds: %#v", rules)
	}
	if rules.TopLevelSliceAllowed != containsSchemaShapeKind(rules.TopLevelPersistentGlobalKinds, SchemaShapeSlice) {
		t.Fatalf("state top-level slice flag disagrees with top-level kinds: %#v", rules)
	}
	if rules.NestedSliceAllowed != containsSchemaShapeKind(rules.NestedStructFieldKinds, SchemaShapeSlice) {
		t.Fatalf("state nested slice flag disagrees with nested field kinds: %#v", rules)
	}
	if rules.SliceScalarAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeScalar) {
		t.Fatalf("state slice scalar flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.SliceNamedStructAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeNamedStruct) {
		t.Fatalf("state slice named-struct flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.AnonymousStructSliceElementsAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeAnonymousStruct) {
		t.Fatalf("state anonymous slice element flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.SliceElementMapsAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeMap) {
		t.Fatalf("state map slice element flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.SliceElementSlicesAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeSlice) {
		t.Fatalf("state slice element flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.MapStringScalarAllowed != (containsString(rules.MapKeyScalarKinds, "string") && containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeScalar)) {
		t.Fatalf("state map string->scalar flag disagrees with map key/value kinds: %#v", rules)
	}
	if rules.MapStringNamedStructAllowed != (containsString(rules.MapKeyScalarKinds, "string") && containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeNamedStruct)) {
		t.Fatalf("state map string->named-struct flag disagrees with map key/value kinds: %#v", rules)
	}
	if rules.AnonymousStructMapValuesAllowed != containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeAnonymousStruct) {
		t.Fatalf("state anonymous map value flag disagrees with map value kinds: %#v", rules)
	}
	if rules.MapValueMapsAllowed != containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeMap) {
		t.Fatalf("state map value map flag disagrees with map value kinds: %#v", rules)
	}
	if rules.MapValueSlicesAllowed != containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeSlice) {
		t.Fatalf("state map value slice flag disagrees with map value kinds: %#v", rules)
	}
}

func assertQueryShapeFlagConsistency(t *testing.T, rules QueryRules) {
	t.Helper()

	if rules.NamedStructAllowed != containsSchemaShapeKind(rules.TopLevelResultKinds, SchemaShapeNamedStruct) {
		t.Fatalf("query named-struct flag disagrees with result kinds: %#v", rules)
	}
	if rules.NestedNamedStructAllowed != containsSchemaShapeKind(rules.NestedStructFieldKinds, SchemaShapeNamedStruct) {
		t.Fatalf("query nested named-struct flag disagrees with nested field kinds: %#v", rules)
	}
	if rules.TopLevelMapAllowed != containsSchemaShapeKind(rules.TopLevelResultKinds, SchemaShapeMap) {
		t.Fatalf("query top-level map flag disagrees with result kinds: %#v", rules)
	}
	if rules.NestedMapAllowed != containsSchemaShapeKind(rules.NestedStructFieldKinds, SchemaShapeMap) {
		t.Fatalf("query nested map flag disagrees with nested field kinds: %#v", rules)
	}
	if rules.TopLevelSliceAllowed != containsSchemaShapeKind(rules.TopLevelResultKinds, SchemaShapeSlice) {
		t.Fatalf("query top-level slice flag disagrees with result kinds: %#v", rules)
	}
	if rules.NestedSliceAllowed != containsSchemaShapeKind(rules.NestedStructFieldKinds, SchemaShapeSlice) {
		t.Fatalf("query nested slice flag disagrees with nested field kinds: %#v", rules)
	}
	if rules.AnonymousResultStructAllowed != containsSchemaShapeKind(rules.TopLevelResultKinds, SchemaShapeAnonymousStruct) {
		t.Fatalf("query anonymous result flag disagrees with result kinds: %#v", rules)
	}
	if !reflect.DeepEqual(rules.BoolPresenceResultValueKinds, rules.TopLevelResultKinds) {
		t.Fatalf("query bool-pair value kinds must match accepted top-level result kinds: %#v", rules)
	}
	if rules.SliceScalarAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeScalar) {
		t.Fatalf("query slice scalar flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.SliceNamedStructAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeNamedStruct) {
		t.Fatalf("query slice named-struct flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.AnonymousStructSliceElementsAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeAnonymousStruct) {
		t.Fatalf("query anonymous slice element flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.SliceElementMapsAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeMap) {
		t.Fatalf("query map slice element flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.SliceElementSlicesAllowed != containsSchemaShapeKind(rules.SliceElementKinds, SchemaShapeSlice) {
		t.Fatalf("query slice element flag disagrees with slice element kinds: %#v", rules)
	}
	if rules.MapStringScalarAllowed != (containsString(rules.MapKeyScalarKinds, "string") && containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeScalar)) {
		t.Fatalf("query map string->scalar flag disagrees with map key/value kinds: %#v", rules)
	}
	if rules.MapStringNamedStructAllowed != (containsString(rules.MapKeyScalarKinds, "string") && containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeNamedStruct)) {
		t.Fatalf("query map string->named-struct flag disagrees with map key/value kinds: %#v", rules)
	}
	if rules.AnonymousStructMapValuesAllowed != containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeAnonymousStruct) {
		t.Fatalf("query anonymous map value flag disagrees with map value kinds: %#v", rules)
	}
	if rules.MapValueMapsAllowed != containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeMap) {
		t.Fatalf("query map value map flag disagrees with map value kinds: %#v", rules)
	}
	if rules.MapValueSlicesAllowed != containsSchemaShapeKind(rules.MapValueKinds, SchemaShapeSlice) {
		t.Fatalf("query map value slice flag disagrees with map value kinds: %#v", rules)
	}
}

func containsSchemaShapeKind(in []SchemaShapeKind, want SchemaShapeKind) bool {
	for _, got := range in {
		if got == want {
			return true
		}
	}
	return false
}

func containsString(in []string, want string) bool {
	for _, got := range in {
		if got == want {
			return true
		}
	}
	return false
}
