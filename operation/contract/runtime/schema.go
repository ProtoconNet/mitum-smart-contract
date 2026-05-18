package runtime

import (
	"fmt"
	"strings"
)

type SchemaMode string

const (
	SchemaModeTypedArgs SchemaMode = "typed-snapshot-v1"
)

const (
	ScalarOnlySupportDescription       = "current Gno typed ABI v1 supports only scalar types: string, bool, int, int64, uint64"
	FlatStructGlobalSupportDescription = "current Gno snapshot v1 supports top-level named struct globals when every field is scalar, []scalar, []named-struct, map[string]scalar, map[string]named-struct, or another named struct; non-string map keys, slice/map elements of unsupported types, anonymous struct fields/elements, pointer/interface fields, and recursive struct types are not supported"
	MapGlobalSupportDescription        = "current Gno snapshot v1 supports maps only when keys are string and values are scalar or named struct trees composed of scalar fields, nested named structs, slices of scalar/named-struct, and map[string]scalar/map[string]named-struct fields; map values cannot be slices or maps directly, and anonymous/recursive struct types are not supported"
	SliceSupportDescription            = "current Gno snapshot v1 supports slices only when elements are scalar or named struct trees composed of scalar fields, nested named structs, slices of scalar/named-struct, and map[string]scalar/map[string]named-struct fields; slice elements cannot be slices, maps, anonymous structs, or recursive named structs"
	QueryResultSupportDescription      = "current Gno query ABI v1 supports T or (T, bool) where T is scalar, named struct, map[string]scalar, map[string]named-struct, []scalar, or []named-struct; named struct fields may use the current snapshot-supported recursive field policy, and anonymous/recursive or otherwise unsupported types are not allowed"
	WriteArgSupportDescription         = "current Gno write ABI v1 supports only scalar parameters: string, bool, int, int64, uint64"
	QueryArgSupportDescription         = "current Gno query ABI v1 supports only scalar parameters: string, bool, int, int64, uint64"
)

var AllowedTypedContractImports = []string{
	MitumChainPackagePath,
	"strconv",
	"strings",
	"errors",
	"bytes",
	"encoding/hex",
	"encoding/base64",
	"unicode/utf8",
}

var allowedTypedContractImportSet = func() map[string]struct{} {
	out := make(map[string]struct{}, len(AllowedTypedContractImports))
	for _, path := range AllowedTypedContractImports {
		out[path] = struct{}{}
	}

	return out
}()

func IsAllowedTypedContractImport(path string) bool {
	_, found := allowedTypedContractImportSet[path]
	return found
}

func AllowedTypedContractImportsDescription() string {
	return strings.Join(AllowedTypedContractImports, ", ")
}

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

	if binding.Type.Kind == TypeSlice {
		return s.validateTopLevelSliceGlobalType(binding)
	}

	if _, err := s.validateNamedStructType(binding.Type, fmt.Sprintf("persistent global %q", binding.Name)); err != nil {
		return fmt.Errorf(
			"%w; %s",
			err,
			FlatStructGlobalSupportDescription,
		)
	}

	return nil
}

func (s ContractSchema) validateTopLevelMapGlobalType(binding PersistentBindingSchema) error {
	if err := s.validateMapTypeRecursive(binding.Type, fmt.Sprintf("persistent global %q", binding.Name), map[string]bool{}); err != nil {
		return fmt.Errorf(
			"%w; %s",
			err,
			MapGlobalSupportDescription,
		)
	}

	return nil
}

func (s ContractSchema) validateTopLevelSliceGlobalType(binding PersistentBindingSchema) error {
	if err := s.validateSliceTypeRecursive(binding.Type, fmt.Sprintf("persistent global %q", binding.Name), map[string]bool{}); err != nil {
		return fmt.Errorf(
			"%w; %s",
			err,
			SliceSupportDescription,
		)
	}

	return nil
}

func (s ContractSchema) validateNamedStructType(typ TypeRef, subject string) (TypeRef, error) {
	return s.validateNamedStructTypeRecursive(typ, subject, map[string]bool{})
}

func (s ContractSchema) validateNamedStructTypeRecursive(typ TypeRef, subject string, resolving map[string]bool) (TypeRef, error) {
	if typ.Kind != TypeNamed {
		return TypeRef{}, fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	resolved, ok := s.Types.Resolve(typ)
	if !ok || resolved.Kind != TypeStruct {
		return TypeRef{}, fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	if resolving[typ.Name] {
		return TypeRef{}, fmt.Errorf(
			"%s type %q is not supported; recursive named struct types are not supported",
			subject,
			typ.String(),
		)
	}

	resolving[typ.Name] = true
	defer delete(resolving, typ.Name)

	for _, field := range resolved.Fields {
		fieldType := s.ResolveType(field.Type)
		switch {
		case fieldType.IsScalar():
			if IsSupportedScalarType(fieldType) {
				continue
			}
			return TypeRef{}, fmt.Errorf(
				"%s field %q type %q is not supported",
				subject,
				field.Name,
				field.Type.String(),
			)

		case field.Type.Kind == TypeNamed:
			if _, err := s.validateNamedStructTypeRecursive(
				field.Type,
				fmt.Sprintf("%s field %q", subject, field.Name),
				resolving,
			); err != nil {
				return TypeRef{}, err
			}

		case field.Type.Kind == TypeStruct:
			return TypeRef{}, fmt.Errorf(
				"%s field %q type %q is not supported; anonymous nested struct fields are not supported",
				subject,
				field.Name,
				field.Type.String(),
			)

		case fieldType.Kind == TypeMap:
			if err := s.validateMapTypeRecursive(
				field.Type,
				fmt.Sprintf("%s field %q", subject, field.Name),
				resolving,
			); err != nil {
				return TypeRef{}, err
			}

		case fieldType.Kind == TypeSlice:
			if err := s.validateSliceTypeRecursive(
				field.Type,
				fmt.Sprintf("%s field %q", subject, field.Name),
				resolving,
			); err != nil {
				return TypeRef{}, err
			}

		default:
			return TypeRef{}, fmt.Errorf(
				"%s field %q type %q is not supported",
				subject,
				field.Name,
				field.Type.String(),
			)
		}
	}

	return resolved, nil
}

func (s ContractSchema) validateSliceTypeRecursive(typ TypeRef, subject string, resolving map[string]bool) error {
	if typ.Kind != TypeSlice || typ.Elem == nil {
		return fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	elemType := s.ResolveType(*typ.Elem)
	switch {
	case elemType.IsScalar():
		if IsSupportedScalarType(elemType) {
			return nil
		}
		return fmt.Errorf("%s slice element type %q is not supported", subject, typ.Elem.String())

	case typ.Elem.Kind == TypeNamed:
		_, err := s.validateNamedStructTypeRecursive(*typ.Elem, fmt.Sprintf("%s slice element", subject), resolving)
		return err

	case typ.Elem.Kind == TypeStruct:
		return fmt.Errorf("%s slice element type %q is not supported; anonymous struct slice elements are not supported", subject, typ.Elem.String())

	case elemType.Kind == TypeMap:
		return fmt.Errorf("%s slice element type %q is not supported; slice elements cannot be maps", subject, typ.Elem.String())

	case elemType.Kind == TypeSlice:
		return fmt.Errorf("%s slice element type %q is not supported; slice elements cannot be slices", subject, typ.Elem.String())

	default:
		return fmt.Errorf("%s slice element type %q is not supported", subject, typ.Elem.String())
	}
}

func (s ContractSchema) validateMapTypeRecursive(typ TypeRef, subject string, resolving map[string]bool) error {
	if typ.Kind != TypeMap || typ.Key == nil || typ.Elem == nil {
		return fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	keyType := s.ResolveType(*typ.Key)
	if !keyType.IsScalar() || keyType.NormalizedString() != "string" {
		return fmt.Errorf("%s map key type %q is not supported", subject, typ.Key.String())
	}

	elemType := s.ResolveType(*typ.Elem)
	switch {
	case elemType.IsScalar():
		if IsSupportedScalarType(elemType) {
			return nil
		}
		return fmt.Errorf("%s map value type %q is not supported", subject, typ.Elem.String())

	case typ.Elem.Kind == TypeNamed:
		_, err := s.validateNamedStructTypeRecursive(*typ.Elem, fmt.Sprintf("%s map value", subject), resolving)
		return err

	case typ.Elem.Kind == TypeStruct:
		return fmt.Errorf("%s map value type %q is not supported; anonymous struct map values are not supported", subject, typ.Elem.String())

	case elemType.Kind == TypeMap:
		return fmt.Errorf("%s map value type %q is not supported; map values cannot be maps", subject, typ.Elem.String())

	case elemType.Kind == TypeSlice:
		return fmt.Errorf("%s map value type %q is not supported; map values cannot be slices", subject, typ.Elem.String())

	default:
		return fmt.Errorf("%s map value type %q is not supported", subject, typ.Elem.String())
	}
}

func (fn FunctionSchema) IsContextCallable() bool {
	if len(fn.Params) < 1 {
		return false
	}

	return isContractContextType(fn.Params[0].Type)
}

func (fn FunctionSchema) IsTypedABIShape() bool {
	if !fn.Exported {
		return false
	}
	if !fn.IsContextCallable() {
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
	return len(fn.Results) == 1 ||
		fn.Results[1].Type.NormalizedString() == "bool"
}

func (s ContractSchema) ValidateQueryResultType(typ TypeRef, subject string) error {
	return s.validateCompositeABIType(typ, subject, QueryResultSupportDescription)
}

func (s ContractSchema) ValidateWriteArgType(typ TypeRef, subject string) error {
	return s.validateScalarABIType(typ, subject, WriteArgSupportDescription)
}

func (s ContractSchema) ValidateQueryArgType(typ TypeRef, subject string) error {
	return s.validateScalarABIType(typ, subject, QueryArgSupportDescription)
}

func (s ContractSchema) validateScalarABIType(typ TypeRef, subject string, description string) error {
	resolved := s.ResolveType(typ)
	if resolved.IsScalar() && IsSupportedScalarType(resolved) {
		return nil
	}

	return fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)
}

func (s ContractSchema) validateCompositeABIType(typ TypeRef, subject string, description string) error {
	resolved := s.ResolveType(typ)

	switch {
	case resolved.IsScalar():
		if IsSupportedScalarType(resolved) {
			return nil
		}
		return fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)

	case typ.Kind == TypeNamed:
		if _, err := s.validateNamedStructType(typ, subject); err != nil {
			return fmt.Errorf("%w; %s", err, description)
		}
		return nil

	case typ.Kind == TypeMap:
		if err := s.validateMapTypeRecursive(typ, subject, map[string]bool{}); err != nil {
			return fmt.Errorf("%w; %s", err, description)
		}
		return nil

	case typ.Kind == TypeSlice:
		if err := s.validateSliceTypeRecursive(typ, subject, map[string]bool{}); err != nil {
			return fmt.Errorf("%w; %s", err, description)
		}
		return nil

	default:
		return fmt.Errorf("%s type %q is not supported; %s", subject, typ.String(), description)
	}
}
