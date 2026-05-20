package runtime

const (
	CurrentSchemaRulesetVersion = "typed-gno-ruleset-v1"
	MaxTypedContractSourceBytes = 256 * 1024
)

type SchemaRuleset struct {
	// Version tracks the semantic contract-schema policy bundle.
	// It is distinct from the snapshot codec version and any future
	// serialized schema metadata format version.
	//
	// Bump this version when accepted/rejected schema meaning changes:
	// package/source rules, import allowlist, scalar kinds, input/state/query
	// shapes, lifecycle semantics, complexity limits, or struct/map/slice
	// constraints. Do not bump for support text wording, error wording,
	// test-only hooks, or implementation refactors that preserve the same
	// schema acceptance policy.
	Version         string
	SourceRules     SourceRules
	ImportRules     ImportRules
	ScalarRules     ScalarRules
	InputRules      InputRules
	StateRules      StateRules
	QueryRules      QueryRules
	LifecycleRules  LifecycleRules
	ComplexityRules ComplexityRules
}

type SourceRules struct {
	RequiredPackageName    string
	MaxContractSourceBytes int
}

type ImportRules struct {
	AllowedImports []string
}

type ScalarRules struct {
	AllowedKinds []string
}

type InputRules struct {
	CompositeInputAllowed bool
}

type SchemaShapeKind string

const (
	SchemaShapeScalar          SchemaShapeKind = "scalar"
	SchemaShapeNamedStruct     SchemaShapeKind = "named-struct"
	SchemaShapeMap             SchemaShapeKind = "map"
	SchemaShapeSlice           SchemaShapeKind = "slice"
	SchemaShapeAnonymousStruct SchemaShapeKind = "anonymous-struct"
)

type StateRules struct {
	TopLevelPersistentGlobalKinds        []SchemaShapeKind
	NestedStructFieldKinds               []SchemaShapeKind
	SliceElementKinds                    []SchemaShapeKind
	MapKeyScalarKinds                    []string
	MapValueKinds                        []SchemaShapeKind
	ExplicitPersistentGlobalTypeRequired bool
	EmbeddedFieldsAllowed                bool
	NamedStructAllowed                   bool
	NestedNamedStructAllowed             bool
	AnonymousNestedStructFieldsAllowed   bool
	RecursiveNamedStructsAllowed         bool
	TopLevelMapAllowed                   bool
	NestedMapAllowed                     bool
	MapStringScalarAllowed               bool
	MapStringNamedStructAllowed          bool
	AnonymousStructMapValuesAllowed      bool
	MapValueSlicesAllowed                bool
	MapValueMapsAllowed                  bool
	TopLevelSliceAllowed                 bool
	NestedSliceAllowed                   bool
	SliceScalarAllowed                   bool
	SliceNamedStructAllowed              bool
	AnonymousStructSliceElementsAllowed  bool
	SliceElementSlicesAllowed            bool
	SliceElementMapsAllowed              bool
}

type QueryRules struct {
	TopLevelResultKinds                 []SchemaShapeKind
	NestedStructFieldKinds              []SchemaShapeKind
	SliceElementKinds                   []SchemaShapeKind
	MapKeyScalarKinds                   []string
	MapValueKinds                       []SchemaShapeKind
	SingleResultAllowed                 bool
	BoolPresenceResultAllowed           bool
	BoolPresenceResultValueKinds        []SchemaShapeKind
	BoolPresenceSecondResultScalarKind  string
	EmbeddedFieldsAllowed               bool
	NamedStructAllowed                  bool
	NestedNamedStructAllowed            bool
	AnonymousResultStructAllowed        bool
	AnonymousNestedStructFieldsAllowed  bool
	RecursiveNamedStructsAllowed        bool
	TopLevelMapAllowed                  bool
	NestedMapAllowed                    bool
	MapStringScalarAllowed              bool
	MapStringNamedStructAllowed         bool
	AnonymousStructMapValuesAllowed     bool
	MapValueSlicesAllowed               bool
	MapValueMapsAllowed                 bool
	TopLevelSliceAllowed                bool
	NestedSliceAllowed                  bool
	SliceScalarAllowed                  bool
	SliceNamedStructAllowed             bool
	AnonymousStructSliceElementsAllowed bool
	SliceElementSlicesAllowed           bool
	SliceElementMapsAllowed             bool
}

type LifecycleRules struct {
	InitializeRegisterOnly bool
	InitializeCallable     bool
}

type ComplexityRules struct {
	MaxImports           int
	MaxFunctions         int
	MaxPersistentGlobals int
	MaxStructs           int
	MaxStructFields      int
	MaxTypeNestingDepth  int
	MaxSchemaNodes       int
}

var currentSchemaRuleset = SchemaRuleset{
	Version: CurrentSchemaRulesetVersion,
	SourceRules: SourceRules{
		RequiredPackageName:    "contract",
		MaxContractSourceBytes: MaxTypedContractSourceBytes,
	},
	ImportRules: ImportRules{
		AllowedImports: []string{
			MitumChainPackagePath,
			"strconv",
			"strings",
			"errors",
			"bytes",
			"encoding/hex",
			"encoding/base64",
			"unicode/utf8",
		},
	},
	ScalarRules: ScalarRules{
		AllowedKinds: []string{
			"string",
			"bool",
			"int",
			"int64",
			"uint64",
		},
	},
	InputRules: InputRules{
		CompositeInputAllowed: false,
	},
	StateRules: StateRules{
		TopLevelPersistentGlobalKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
			SchemaShapeMap,
			SchemaShapeSlice,
		},
		NestedStructFieldKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
			SchemaShapeMap,
			SchemaShapeSlice,
		},
		SliceElementKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
		},
		MapKeyScalarKinds: []string{
			"string",
		},
		MapValueKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
		},
		ExplicitPersistentGlobalTypeRequired: true,
		EmbeddedFieldsAllowed:                false,
		NamedStructAllowed:                   true,
		NestedNamedStructAllowed:             true,
		AnonymousNestedStructFieldsAllowed:   false,
		RecursiveNamedStructsAllowed:         false,
		TopLevelMapAllowed:                   true,
		NestedMapAllowed:                     true,
		MapStringScalarAllowed:               true,
		MapStringNamedStructAllowed:          true,
		AnonymousStructMapValuesAllowed:      false,
		MapValueSlicesAllowed:                false,
		MapValueMapsAllowed:                  false,
		TopLevelSliceAllowed:                 true,
		NestedSliceAllowed:                   true,
		SliceScalarAllowed:                   true,
		SliceNamedStructAllowed:              true,
		AnonymousStructSliceElementsAllowed:  false,
		SliceElementSlicesAllowed:            false,
		SliceElementMapsAllowed:              false,
	},
	QueryRules: QueryRules{
		TopLevelResultKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
			SchemaShapeMap,
			SchemaShapeSlice,
		},
		NestedStructFieldKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
			SchemaShapeMap,
			SchemaShapeSlice,
		},
		SliceElementKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
		},
		MapKeyScalarKinds: []string{
			"string",
		},
		MapValueKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
		},
		SingleResultAllowed:       true,
		BoolPresenceResultAllowed: true,
		BoolPresenceResultValueKinds: []SchemaShapeKind{
			SchemaShapeScalar,
			SchemaShapeNamedStruct,
			SchemaShapeMap,
			SchemaShapeSlice,
		},
		BoolPresenceSecondResultScalarKind:  "bool",
		EmbeddedFieldsAllowed:               false,
		NamedStructAllowed:                  true,
		NestedNamedStructAllowed:            true,
		AnonymousResultStructAllowed:        false,
		AnonymousNestedStructFieldsAllowed:  false,
		RecursiveNamedStructsAllowed:        false,
		TopLevelMapAllowed:                  true,
		NestedMapAllowed:                    true,
		MapStringScalarAllowed:              true,
		MapStringNamedStructAllowed:         true,
		AnonymousStructMapValuesAllowed:     false,
		MapValueSlicesAllowed:               false,
		MapValueMapsAllowed:                 false,
		TopLevelSliceAllowed:                true,
		NestedSliceAllowed:                  true,
		SliceScalarAllowed:                  true,
		SliceNamedStructAllowed:             true,
		AnonymousStructSliceElementsAllowed: false,
		SliceElementSlicesAllowed:           false,
		SliceElementMapsAllowed:             false,
	},
	LifecycleRules: LifecycleRules{
		InitializeRegisterOnly: true,
		InitializeCallable:     false,
	},
	ComplexityRules: ComplexityRules{
		MaxImports:           MaxContractSchemaImports,
		MaxFunctions:         MaxContractSchemaFunctions,
		MaxPersistentGlobals: MaxContractSchemaPersistentGlobals,
		MaxStructs:           MaxContractSchemaStructs,
		MaxStructFields:      MaxContractSchemaStructFields,
		MaxTypeNestingDepth:  MaxContractSchemaTypeNestingDepth,
		MaxSchemaNodes:       MaxContractSchemaNodes,
	},
}

func CurrentSchemaRuleset() SchemaRuleset {
	return currentSchemaRuleset.clone()
}

func (r SchemaRuleset) clone() SchemaRuleset {
	r.SourceRules = r.SourceRules
	r.ImportRules = r.ImportRules.clone()
	r.ScalarRules = r.ScalarRules.clone()
	r.InputRules = r.InputRules.clone()
	r.StateRules = r.StateRules.clone()
	r.QueryRules = r.QueryRules.clone()
	r.LifecycleRules = r.LifecycleRules
	r.ComplexityRules = r.ComplexityRules

	return r
}

func (r ImportRules) clone() ImportRules {
	r.AllowedImports = cloneStringSlice(r.AllowedImports)
	return r
}

func (r ScalarRules) clone() ScalarRules {
	r.AllowedKinds = cloneStringSlice(r.AllowedKinds)
	return r
}

func (r InputRules) clone() InputRules { return r }

func (r StateRules) clone() StateRules {
	r.TopLevelPersistentGlobalKinds = cloneSchemaShapeKindSlice(r.TopLevelPersistentGlobalKinds)
	r.NestedStructFieldKinds = cloneSchemaShapeKindSlice(r.NestedStructFieldKinds)
	r.SliceElementKinds = cloneSchemaShapeKindSlice(r.SliceElementKinds)
	r.MapKeyScalarKinds = cloneStringSlice(r.MapKeyScalarKinds)
	r.MapValueKinds = cloneSchemaShapeKindSlice(r.MapValueKinds)
	return r
}

func (r QueryRules) clone() QueryRules {
	r.TopLevelResultKinds = cloneSchemaShapeKindSlice(r.TopLevelResultKinds)
	r.NestedStructFieldKinds = cloneSchemaShapeKindSlice(r.NestedStructFieldKinds)
	r.SliceElementKinds = cloneSchemaShapeKindSlice(r.SliceElementKinds)
	r.MapKeyScalarKinds = cloneStringSlice(r.MapKeyScalarKinds)
	r.MapValueKinds = cloneSchemaShapeKindSlice(r.MapValueKinds)
	r.BoolPresenceResultValueKinds = cloneSchemaShapeKindSlice(r.BoolPresenceResultValueKinds)
	return r
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}

	out := make([]string, len(in))
	copy(out, in)

	return out
}

func cloneSchemaShapeKindSlice(in []SchemaShapeKind) []SchemaShapeKind {
	if len(in) == 0 {
		return nil
	}

	out := make([]SchemaShapeKind, len(in))
	copy(out, in)

	return out
}
