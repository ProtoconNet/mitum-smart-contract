package runtime

import (
	"fmt"

	contracttypes "github.com/ProtoconNet/mitum-smart-contract/types"
)

func NewPersistedContractSchema(
	sourceCode string,
	schema ContractSchema,
) contracttypes.PersistedContractSchema {
	return contracttypes.PersistedContractSchema{
		SchemaFormatVersion:  contracttypes.CurrentSchemaFormatVersion,
		SchemaRulesetVersion: CurrentSchemaRulesetVersion,
		SourceHash:           contracttypes.ContractSourceHash(sourceCode),
		Schema:               persistedContractSchemaFromRuntime(schema),
	}
}

func RuntimeSchemaFromPersisted(
	sourceCode string,
	persisted *contracttypes.PersistedContractSchema,
) (ContractSchema, bool) {
	if persisted == nil {
		return ContractSchema{}, false
	}
	if err := persisted.IsValid(nil); err != nil {
		return ContractSchema{}, false
	}
	if persisted.SchemaFormatVersion != contracttypes.CurrentSchemaFormatVersion {
		return ContractSchema{}, false
	}
	if persisted.SchemaRulesetVersion != CurrentSchemaRulesetVersion {
		return ContractSchema{}, false
	}
	if persisted.SourceHash != contracttypes.ContractSourceHash(sourceCode) {
		return ContractSchema{}, false
	}
	if persisted.Schema.Mode != string(SchemaModeTypedArgs) {
		return ContractSchema{}, false
	}

	schema, err := runtimeContractSchemaFromPersisted(persisted.Schema)
	if err != nil {
		return ContractSchema{}, false
	}
	if err := finalizeContractSchema(&schema); err != nil {
		return ContractSchema{}, false
	}
	if schema.Mode != SchemaModeTypedArgs {
		return ContractSchema{}, false
	}

	storeContractSchemaInCache(sourceCode, schema)

	return schema, true
}

func persistedContractSchemaFromRuntime(schema ContractSchema) contracttypes.ContractSchema {
	return contracttypes.ContractSchema{
		PackageName:       schema.PackageName,
		Mode:              string(schema.Mode),
		Types:             persistedTypeRegistryFromRuntime(schema.Types),
		PersistentGlobals: persistedPersistentBindingSchemasFromRuntime(schema.PersistentGlobals),
		Functions:         persistedFunctionSchemasFromRuntime(schema.Functions),
	}
}

func persistedTypeRegistryFromRuntime(reg TypeRegistry) contracttypes.TypeRegistry {
	if len(reg.Structs) == 0 {
		return contracttypes.TypeRegistry{}
	}

	structs := make(map[string]contracttypes.TypeRef, len(reg.Structs))
	for name, typ := range reg.Structs {
		structs[name] = persistedTypeRefFromRuntime(typ)
	}

	return contracttypes.TypeRegistry{Structs: structs}
}

func persistedTypeRefFromRuntime(typ TypeRef) contracttypes.TypeRef {
	out := contracttypes.TypeRef{
		Kind:   string(typ.Kind),
		Raw:    typ.Raw,
		Scalar: typ.Scalar,
		Name:   typ.Name,
		Fields: persistedStructFieldsFromRuntime(typ.Fields),
	}
	if typ.Key != nil {
		key := persistedTypeRefFromRuntime(*typ.Key)
		out.Key = &key
	}
	if typ.Elem != nil {
		elem := persistedTypeRefFromRuntime(*typ.Elem)
		out.Elem = &elem
	}

	return out
}

func persistedStructFieldsFromRuntime(in []StructField) []contracttypes.StructField {
	if len(in) == 0 {
		return nil
	}

	out := make([]contracttypes.StructField, len(in))
	for i := range in {
		out[i] = contracttypes.StructField{
			Name: in[i].Name,
			Type: persistedTypeRefFromRuntime(in[i].Type),
		}
	}

	return out
}

func persistedPersistentBindingSchemasFromRuntime(in []PersistentBindingSchema) []contracttypes.PersistentBindingSchema {
	if len(in) == 0 {
		return nil
	}

	out := make([]contracttypes.PersistentBindingSchema, len(in))
	for i := range in {
		out[i] = contracttypes.PersistentBindingSchema{
			Name:            in[i].Name,
			Type:            persistedTypeRefFromRuntime(in[i].Type),
			HasExplicitType: in[i].HasExplicitType,
		}
	}

	return out
}

func persistedFunctionSchemasFromRuntime(in []FunctionSchema) []contracttypes.FunctionSchema {
	if len(in) == 0 {
		return nil
	}

	out := make([]contracttypes.FunctionSchema, len(in))
	for i := range in {
		out[i] = contracttypes.FunctionSchema{
			Name:     in[i].Name,
			Exported: in[i].Exported,
			Params:   persistedParamSchemasFromRuntime(in[i].Params),
			Results:  persistedResultSchemasFromRuntime(in[i].Results),
		}
	}

	return out
}

func persistedParamSchemasFromRuntime(in []ParamSchema) []contracttypes.ParamSchema {
	if len(in) == 0 {
		return nil
	}

	out := make([]contracttypes.ParamSchema, len(in))
	for i := range in {
		out[i] = contracttypes.ParamSchema{
			Name: in[i].Name,
			Type: persistedTypeRefFromRuntime(in[i].Type),
		}
	}

	return out
}

func persistedResultSchemasFromRuntime(in []ResultSchema) []contracttypes.ResultSchema {
	if len(in) == 0 {
		return nil
	}

	out := make([]contracttypes.ResultSchema, len(in))
	for i := range in {
		out[i] = contracttypes.ResultSchema{
			Type: persistedTypeRefFromRuntime(in[i].Type),
		}
	}

	return out
}

func runtimeContractSchemaFromPersisted(dto contracttypes.ContractSchema) (ContractSchema, error) {
	if dto.Mode != string(SchemaModeTypedArgs) {
		return ContractSchema{}, fmt.Errorf("unsupported schema mode %q", dto.Mode)
	}

	reg, err := runtimeTypeRegistryFromPersisted(dto.Types)
	if err != nil {
		return ContractSchema{}, err
	}

	globals, err := runtimePersistentBindingSchemasFromPersisted(dto.PersistentGlobals)
	if err != nil {
		return ContractSchema{}, err
	}

	functions, err := runtimeFunctionSchemasFromPersisted(dto.Functions)
	if err != nil {
		return ContractSchema{}, err
	}

	return ContractSchema{
		PackageName:       dto.PackageName,
		Mode:              SchemaMode(dto.Mode),
		Types:             reg,
		PersistentGlobals: globals,
		Functions:         functions,
	}, nil
}

func runtimeTypeRegistryFromPersisted(dto contracttypes.TypeRegistry) (TypeRegistry, error) {
	out := NewTypeRegistry()
	for name, typ := range dto.Structs {
		rt, err := runtimeTypeRefFromPersisted(typ)
		if err != nil {
			return TypeRegistry{}, fmt.Errorf("struct %q: %w", name, err)
		}
		out.Structs[name] = rt
	}

	return out, nil
}

func runtimeTypeRefFromPersisted(dto contracttypes.TypeRef) (TypeRef, error) {
	kind := TypeKind(dto.Kind)
	if !isKnownTypeKind(kind) {
		return TypeRef{}, fmt.Errorf("unsupported type kind %q", dto.Kind)
	}

	if err := validatePersistedTypeRefShape(dto, kind); err != nil {
		return TypeRef{}, err
	}

	out := TypeRef{
		Kind:   kind,
		Raw:    dto.Raw,
		Scalar: dto.Scalar,
		Name:   dto.Name,
	}

	if dto.Key != nil {
		key, err := runtimeTypeRefFromPersisted(*dto.Key)
		if err != nil {
			return TypeRef{}, fmt.Errorf("map key: %w", err)
		}
		out.Key = &key
	}
	if dto.Elem != nil {
		elem, err := runtimeTypeRefFromPersisted(*dto.Elem)
		if err != nil {
			return TypeRef{}, fmt.Errorf("element: %w", err)
		}
		out.Elem = &elem
	}
	if len(dto.Fields) > 0 {
		fields, err := runtimeStructFieldsFromPersisted(dto.Fields)
		if err != nil {
			return TypeRef{}, err
		}
		out.Fields = fields
	}

	return out, nil
}

func isKnownTypeKind(kind TypeKind) bool {
	switch kind {
	case TypeScalar, TypeStruct, TypeMap, TypeSlice, TypeNamed, TypeOpaque:
		return true
	default:
		return false
	}
}

func validatePersistedTypeRefShape(dto contracttypes.TypeRef, kind TypeKind) error {
	switch kind {
	case TypeMap:
		if dto.Key == nil || dto.Elem == nil {
			return fmt.Errorf("map type requires key and element")
		}
		if len(dto.Fields) > 0 {
			return fmt.Errorf("map type cannot have fields")
		}
	case TypeSlice:
		if dto.Elem == nil {
			return fmt.Errorf("slice type requires element")
		}
		if dto.Key != nil || len(dto.Fields) > 0 {
			return fmt.Errorf("slice type cannot have key or fields")
		}
	case TypeStruct:
		if dto.Key != nil || dto.Elem != nil {
			return fmt.Errorf("struct type cannot have key or element")
		}
	case TypeScalar, TypeNamed, TypeOpaque:
		if dto.Key != nil || dto.Elem != nil || len(dto.Fields) > 0 {
			return fmt.Errorf("%s type cannot have key, element, or fields", kind)
		}
	}

	return nil
}

func runtimeStructFieldsFromPersisted(in []contracttypes.StructField) ([]StructField, error) {
	out := make([]StructField, len(in))
	for i := range in {
		typ, err := runtimeTypeRefFromPersisted(in[i].Type)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", in[i].Name, err)
		}
		out[i] = StructField{
			Name: in[i].Name,
			Type: typ,
		}
	}

	return out, nil
}

func runtimePersistentBindingSchemasFromPersisted(in []contracttypes.PersistentBindingSchema) ([]PersistentBindingSchema, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]PersistentBindingSchema, len(in))
	for i := range in {
		typ, err := runtimeTypeRefFromPersisted(in[i].Type)
		if err != nil {
			return nil, fmt.Errorf("persistent global %q: %w", in[i].Name, err)
		}
		out[i] = PersistentBindingSchema{
			Name:            in[i].Name,
			Type:            typ,
			HasExplicitType: in[i].HasExplicitType,
		}
	}

	return out, nil
}

func runtimeFunctionSchemasFromPersisted(in []contracttypes.FunctionSchema) ([]FunctionSchema, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]FunctionSchema, len(in))
	for i := range in {
		params, err := runtimeParamSchemasFromPersisted(in[i].Params)
		if err != nil {
			return nil, fmt.Errorf("function %q params: %w", in[i].Name, err)
		}
		results, err := runtimeResultSchemasFromPersisted(in[i].Results)
		if err != nil {
			return nil, fmt.Errorf("function %q results: %w", in[i].Name, err)
		}
		out[i] = FunctionSchema{
			Name:     in[i].Name,
			Exported: in[i].Exported,
			Params:   params,
			Results:  results,
		}
	}

	return out, nil
}

func runtimeParamSchemasFromPersisted(in []contracttypes.ParamSchema) ([]ParamSchema, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]ParamSchema, len(in))
	for i := range in {
		typ, err := runtimeTypeRefFromPersisted(in[i].Type)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", in[i].Name, err)
		}
		out[i] = ParamSchema{
			Name: in[i].Name,
			Type: typ,
		}
	}

	return out, nil
}

func runtimeResultSchemasFromPersisted(in []contracttypes.ResultSchema) ([]ResultSchema, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make([]ResultSchema, len(in))
	for i := range in {
		typ, err := runtimeTypeRefFromPersisted(in[i].Type)
		if err != nil {
			return nil, fmt.Errorf("result %d: %w", i, err)
		}
		out[i] = ResultSchema{Type: typ}
	}

	return out, nil
}
