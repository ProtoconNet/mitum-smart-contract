package runtime

import (
	"fmt"
	"strings"
)

type SchemaMode string

const (
	SchemaModeLegacyMap SchemaMode = "legacy-map-v1"
	SchemaModeTypedArgs SchemaMode = "typed-snapshot-v1"
)

const (
	ScalarOnlySupportDescription       = "current Gno typed ABI v1 supports only scalar types: string, bool, int, int64, uint64"
	FlatStructGlobalSupportDescription = "current Gno snapshot v1 supports scalar globals and top-level named struct globals with scalar fields only"
	MapGlobalSupportDescription        = "current Gno snapshot v1 supports top-level map[string]scalar globals and top-level map[string]flat-struct globals only; map keys must be string and map values must be scalar or named struct with scalar fields only"
)

type TypeKind string

const (
	TypeScalar TypeKind = "scalar"
	TypeStruct TypeKind = "struct"
	TypeMap    TypeKind = "map"
	TypeSlice  TypeKind = "slice"
	TypeNamed  TypeKind = "named"
	TypeOpaque TypeKind = "opaque"
)

type TypeRef struct {
	Kind   TypeKind
	Raw    string
	Scalar string
	Name   string
	Key    *TypeRef
	Elem   *TypeRef
	Fields []StructField
}

type StructField struct {
	Name string
	Type TypeRef
}

type TypeRegistry struct {
	Structs map[string]TypeRef
}

func NewTypeRegistry() TypeRegistry {
	return TypeRegistry{
		Structs: map[string]TypeRef{},
	}
}

func (r TypeRegistry) Resolve(t TypeRef) (TypeRef, bool) {
	if t.Kind != TypeNamed {
		return t, true
	}

	resolved, found := r.Structs[t.Name]
	if !found {
		return t, false
	}

	return resolved, true
}

type ContractSchema struct {
	PackageName       string
	Mode              SchemaMode
	Types             TypeRegistry
	PersistentGlobals []PersistentBindingSchema
	Functions         []FunctionSchema
}

type PersistentBindingSchema struct {
	Name            string
	Type            TypeRef
	HasExplicitType bool
}

type FunctionSchema struct {
	Name     string
	Exported bool
	Params   []ParamSchema
	Results  []ResultSchema
}

type ParamSchema struct {
	Name string
	Type TypeRef
}

type ResultSchema struct {
	Type TypeRef
}

func (t TypeRef) String() string {
	if t.Raw != "" {
		return t.Raw
	}

	switch t.Kind {
	case TypeScalar:
		return t.Scalar
	case TypeNamed:
		return t.Name
	case TypeMap:
		if t.Key == nil || t.Elem == nil {
			return "map[?]?"
		}
		return "map[" + t.Key.String() + "]" + t.Elem.String()
	case TypeSlice:
		if t.Elem == nil {
			return "[]?"
		}
		return "[]" + t.Elem.String()
	case TypeStruct:
		if t.Name != "" {
			return t.Name
		}
		var b strings.Builder
		b.WriteString("struct{")
		for i, field := range t.Fields {
			if i > 0 {
				b.WriteString(";")
			}
			b.WriteString(field.Name)
			b.WriteString(" ")
			b.WriteString(field.Type.String())
		}
		b.WriteString("}")
		return b.String()
	default:
		return ""
	}
}

func (t TypeRef) NormalizedString() string {
	return normalizeTypeString(t.String())
}

func (t TypeRef) IsScalar() bool {
	return t.Kind == TypeScalar
}

func (s ContractSchema) FindFunction(name string) (FunctionSchema, bool) {
	for _, fn := range s.Functions {
		if fn.Name == name {
			return fn, true
		}
	}

	return FunctionSchema{}, false
}

func (s ContractSchema) ResolveType(t TypeRef) TypeRef {
	resolved, ok := s.Types.Resolve(t)
	if !ok {
		return t
	}

	return resolved
}

func (s ContractSchema) SupportsScalarOnlyType(t TypeRef) bool {
	return s.ResolveType(t).IsScalar()
}

func (s ContractSchema) ValidatePersistentGlobalType(binding PersistentBindingSchema) error {
	if s.SupportsScalarOnlyType(binding.Type) {
		return nil
	}

	if binding.Type.Kind == TypeMap {
		return s.validateTopLevelMapGlobalType(binding)
	}

	if _, err := s.validateFlatNamedStructType(binding.Type, fmt.Sprintf("persistent global %q", binding.Name)); err != nil {
		return fmt.Errorf(
			"%w; %s",
			err,
			FlatStructGlobalSupportDescription,
		)
	}

	return nil
}

func (s ContractSchema) validateTopLevelMapGlobalType(binding PersistentBindingSchema) error {
	if binding.Type.Key == nil || binding.Type.Elem == nil {
		return fmt.Errorf(
			"persistent global %q type %q is not supported; %s",
			binding.Name,
			binding.Type.String(),
			MapGlobalSupportDescription,
		)
	}

	keyType := s.ResolveType(*binding.Type.Key)
	if !keyType.IsScalar() || keyType.NormalizedString() != "string" {
		return fmt.Errorf(
			"persistent global %q map key type %q is not supported; %s",
			binding.Name,
			binding.Type.Key.String(),
			MapGlobalSupportDescription,
		)
	}

	elemType := s.ResolveType(*binding.Type.Elem)
	if elemType.IsScalar() && IsSupportedScalarType(elemType) {
		return nil
	}

	if _, err := s.validateFlatNamedStructType(*binding.Type.Elem, fmt.Sprintf("persistent global %q map value", binding.Name)); err != nil {
		return fmt.Errorf(
			"%w; %s",
			err,
			MapGlobalSupportDescription,
		)
	}

	return nil
}

func (s ContractSchema) validateFlatNamedStructType(typ TypeRef, subject string) (TypeRef, error) {
	if typ.Kind != TypeNamed {
		return TypeRef{}, fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	resolved, ok := s.Types.Resolve(typ)
	if !ok || resolved.Kind != TypeStruct {
		return TypeRef{}, fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	for _, field := range resolved.Fields {
		fieldType := s.ResolveType(field.Type)
		if !fieldType.IsScalar() || !IsSupportedScalarType(fieldType) {
			return TypeRef{}, fmt.Errorf(
				"%s field %q type %q is not supported; flat struct globals require scalar fields only",
				subject,
				field.Name,
				field.Type.String(),
			)
		}
	}

	return resolved, nil
}

func (fn FunctionSchema) IsContextCallable() bool {
	if len(fn.Params) < 1 {
		return false
	}

	return isContractContextType(fn.Params[0].Type)
}

func (fn FunctionSchema) IsLegacyABIShape() bool {
	if !fn.IsContextCallable() {
		return false
	}

	if len(fn.Params) != 1 {
		return false
	}

	if len(fn.Results) != 2 {
		return false
	}

	return fn.Results[0].Type.NormalizedString() == "map[string]interface{}" &&
		fn.Results[1].Type.NormalizedString() == "error"
}

func (fn FunctionSchema) IsTypedABIShape() bool {
	if !fn.Exported {
		return false
	}
	if !fn.IsContextCallable() {
		return false
	}
	if fn.IsLegacyABIShape() {
		return false
	}
	if len(fn.Params) < 2 {
		return false
	}

	return true
}

func isContractContextType(typ TypeRef) bool {
	t := typ.NormalizedString()

	return t == "ContractContext" || strings.HasSuffix(t, ".ContractContext")
}

func normalizeTypeString(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")

	return s
}

func (fn FunctionSchema) IsSingleErrorResult() bool {
	return len(fn.Results) == 1 &&
		fn.Results[0].Type.NormalizedString() == "error"
}

func (fn FunctionSchema) IsTypedInitializeShape() bool {
	return fn.Name == "Initialize" &&
		fn.IsContextCallable() &&
		len(fn.Params) == 1 &&
		fn.IsSingleErrorResult()
}

func (fn FunctionSchema) IsTypedWriteShape() bool {
	return fn.Exported &&
		fn.Name != "Initialize" &&
		fn.IsContextCallable() &&
		len(fn.Params) >= 2 &&
		fn.IsSingleErrorResult()
}

func IsSupportedScalarType(typ TypeRef) bool {
	switch typ.Kind {
	case TypeScalar:
		switch typ.NormalizedString() {
		case "string", "bool", "int", "int64", "uint64":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func IsSupportedScalarResultType(typ TypeRef) bool {
	return IsSupportedScalarType(typ)
}

func (fn FunctionSchema) IsTypedQueryShape() bool {
	if !fn.Exported {
		return false
	}
	if fn.Name == "Initialize" {
		return false
	}
	if !fn.IsContextCallable() {
		return false
	}
	if len(fn.Results) < 1 || len(fn.Results) > 2 {
		return false
	}
	if len(fn.Results) == 1 {
		return IsSupportedScalarResultType(fn.Results[0].Type)
	}

	return IsSupportedScalarResultType(fn.Results[0].Type) &&
		fn.Results[1].Type.NormalizedString() == "bool"
}
