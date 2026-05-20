package runtime

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

const (
	GnoSnapshotCodecName = "gno-snapshot-json-v1"
	GnoSnapshotVersion   = uint64(1)
)

type SnapshotDoc struct {
	Version  uint64            `json:"version"`
	Bindings []SnapshotBinding `json:"bindings"`
}

type SnapshotBinding struct {
	Name  string        `json:"name"`
	Value SnapshotValue `json:"value"`
}

type SnapshotValue struct {
	Kind    string             `json:"kind"`
	Scalar  string             `json:"scalar,omitempty"`
	Fields  []SnapshotField    `json:"fields,omitempty"`
	Entries []SnapshotMapEntry `json:"entries,omitempty"`
	Items   []SnapshotValue    `json:"items,omitempty"`
	IsNil   bool               `json:"is_nil,omitempty"`
}

type SnapshotField struct {
	Name  string        `json:"name"`
	Value SnapshotValue `json:"value"`
}

type SnapshotMapEntry struct {
	Key   string        `json:"key"`
	Value SnapshotValue `json:"value"`
}

func CaptureSnapshot(pkg *gno.PackageValue, store gno.Store, schema ContractSchema) ([]byte, error) {
	doc := SnapshotDoc{
		Version: GnoSnapshotVersion,
	}

	pn := pkg.GetPackageNode(store)
	pb := pkg.GetBlock(store)

	for _, binding := range schema.PersistentGlobals {
		path := pn.GetPathForName(store, gno.Name(binding.Name))
		ptr := pb.GetPointerTo(store, path)
		deepFillSnapshotTypedValue(schema, binding.Type, ptr.TV, store)

		value, err := ExtractSnapshotValue(schema, binding.Type, *ptr.TV)
		if err != nil {
			return nil, fmt.Errorf("capture %s: %w", binding.Name, err)
		}

		doc.Bindings = append(doc.Bindings, SnapshotBinding{
			Name:  binding.Name,
			Value: value,
		})
	}

	snapshotBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	if err := ValidateSnapshotLimits(doc, snapshotBytes); err != nil {
		return nil, err
	}

	return snapshotBytes, nil
}

func RestoreSnapshot(m *gno.Machine, pkg *gno.PackageValue, snapshot []byte, schema ContractSchema) error {
	if len(snapshot) == 0 {
		return nil
	}
	if err := validateSnapshotSizeLimit(snapshot); err != nil {
		return err
	}

	var doc SnapshotDoc
	if err := json.Unmarshal(snapshot, &doc); err != nil {
		return err
	}
	if err := ValidateSnapshotLimits(doc, snapshot); err != nil {
		return err
	}

	m.SetActivePackage(pkg)
	m.PushBlock(pkg.GetBlock(m.Store))
	defer m.PopBlock()

	for _, binding := range doc.Bindings {
		schemaBinding, found := findPersistentBinding(schema, binding.Name)
		if !found {
			return fmt.Errorf("restore %s: binding not found in schema", binding.Name)
		}

		expr, err := BuildExpr(schema, schemaBinding.Type, binding.Value)
		if err != nil {
			return fmt.Errorf("restore %s: %w", binding.Name, err)
		}

		m.RunStatement(gno.StageRun, gno.A(binding.Name, "=", expr))
	}

	return nil
}

func findPersistentBinding(schema ContractSchema, name string) (PersistentBindingSchema, bool) {
	for _, binding := range schema.PersistentGlobals {
		if binding.Name == name {
			return binding, true
		}
	}

	return PersistentBindingSchema{}, false
}

func ExtractSnapshotValue(schema ContractSchema, typ TypeRef, tv gno.TypedValue) (SnapshotValue, error) {
	resolved := schema.ResolveType(typ)

	switch resolved.Kind {
	case TypeScalar:
		value, err := scalarTypedValueToString(resolved, tv)
		if err != nil {
			return SnapshotValue{}, err
		}

		return SnapshotValue{
			Kind:   string(TypeScalar),
			Scalar: value,
		}, nil

	case TypeStruct:
		if typ.Kind != TypeNamed {
			return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), FlatStructGlobalSupportDescription)
		}

		return extractStructSnapshotValue(schema, typ, resolved, tv)

	case TypeMap:
		return extractMapSnapshotValue(schema, typ, resolved, tv)

	case TypeSlice:
		return extractSliceSnapshotValue(schema, typ, resolved, tv)

	default:
		return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), MapGlobalSupportDescription)
	}
}

func BuildLiteral(schema ContractSchema, typ TypeRef, sv SnapshotValue) (string, error) {
	resolved := schema.ResolveType(typ)

	switch resolved.Kind {
	case TypeScalar:
		return scalarLiteral(resolved, sv)

	case TypeStruct:
		if typ.Kind != TypeNamed {
			return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), FlatStructGlobalSupportDescription)
		}

		return buildStructLiteral(schema, typ, resolved, sv)

	case TypeMap:
		return buildMapLiteral(schema, typ, resolved, sv)

	case TypeSlice:
		return buildSliceLiteral(schema, typ, resolved, sv)

	default:
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}
}

func BuildExpr(schema ContractSchema, typ TypeRef, sv SnapshotValue) (gno.Expr, error) {
	resolved := schema.ResolveType(typ)

	switch resolved.Kind {
	case TypeScalar:
		return scalarExpr(resolved, sv)

	case TypeStruct:
		if typ.Kind != TypeNamed {
			return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), FlatStructGlobalSupportDescription)
		}

		return buildStructExpr(schema, typ, resolved, sv)

	case TypeMap:
		return buildMapExpr(schema, typ, resolved, sv)

	case TypeSlice:
		return buildSliceExpr(schema, typ, resolved, sv)

	default:
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}
}

func deepFillSnapshotTypedValue(schema ContractSchema, typ TypeRef, tv *gno.TypedValue, store gno.Store) {
	if tv == nil {
		return
	}

	resolved := schema.ResolveType(typ)
	switch resolved.Kind {
	case TypeMap:
		if tv.V == nil {
			return
		}
		mv, ok := tv.V.(*gno.MapValue)
		if !ok || mv == nil || mv.List == nil {
			return
		}
		for item := mv.List.Head; item != nil; item = item.Next {
			item.Key.DeepFill(store)
			deepFillSnapshotTypedValue(schema, *resolved.Elem, &item.Value, store)
		}
	case TypeSlice:
		if tv.V == nil {
			return
		}
		sv, ok := tv.V.(*gno.SliceValue)
		if !ok || sv == nil {
			tv.DeepFill(store)
			return
		}
		base := sv.GetBase(store)
		if base == nil {
			return
		}
		for i := 0; i < sv.Length; i++ {
			deepFillSnapshotTypedValue(schema, *resolved.Elem, &base.List[sv.Offset+i], store)
		}
	case TypeStruct:
		structValue, ok := tv.V.(*gno.StructValue)
		if !ok || structValue == nil {
			tv.DeepFill(store)
			return
		}
		for i := range resolved.Fields {
			if i >= len(structValue.Fields) {
				break
			}
			deepFillSnapshotTypedValue(schema, resolved.Fields[i].Type, &structValue.Fields[i], store)
		}
	default:
		tv.DeepFill(store)
	}
}

func extractStructSnapshotValue(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	tv gno.TypedValue,
) (SnapshotValue, error) {
	structValue, ok := tv.V.(*gno.StructValue)
	if !ok || structValue == nil {
		return SnapshotValue{}, fmt.Errorf("expected struct value for %q", typ.String())
	}
	if len(structValue.Fields) != len(resolved.Fields) {
		return SnapshotValue{}, fmt.Errorf(
			"struct field count mismatch for %q: expected %d, got %d",
			typ.String(),
			len(resolved.Fields),
			len(structValue.Fields),
		)
	}

	fields := make([]SnapshotField, 0, len(resolved.Fields))
	for i, field := range resolved.Fields {
		if err := validateStructFieldSnapshotType(schema, field.Type, typ.String(), field.Name); err != nil {
			return SnapshotValue{}, err
		}

		value, err := ExtractSnapshotValue(schema, field.Type, structValue.Fields[i])
		if err != nil {
			return SnapshotValue{}, err
		}

		fields = append(fields, SnapshotField{
			Name:  field.Name,
			Value: value,
		})
	}

	return SnapshotValue{
		Kind:   string(TypeStruct),
		Fields: fields,
	}, nil
}

func buildStructLiteral(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	sv SnapshotValue,
) (string, error) {
	if sv.Kind != string(TypeStruct) {
		return "", fmt.Errorf("expected struct snapshot value for type %q", typ.String())
	}
	if len(sv.Fields) != len(resolved.Fields) {
		return "", fmt.Errorf(
			"expected %d struct fields for %q, got %d",
			len(resolved.Fields),
			typ.String(),
			len(sv.Fields),
		)
	}

	structName := typ.Name
	if structName == "" {
		structName = resolved.Name
	}
	if structName == "" {
		return "", fmt.Errorf("struct literal for %q requires a named struct type", typ.String())
	}

	fieldLiterals := make([]string, 0, len(resolved.Fields))
	for i, field := range resolved.Fields {
		snapshotField := sv.Fields[i]
		if snapshotField.Name != field.Name {
			return "", fmt.Errorf(
				"expected struct field %q at index %d for %q, got %q",
				field.Name,
				i,
				typ.String(),
				snapshotField.Name,
			)
		}

		if err := validateStructFieldSnapshotType(schema, field.Type, typ.String(), field.Name); err != nil {
			return "", err
		}

		lit, err := BuildLiteral(schema, field.Type, snapshotField.Value)
		if err != nil {
			return "", err
		}

		fieldLiterals = append(fieldLiterals, field.Name+":"+lit)
	}

	return structName + "{" + joinWithComma(fieldLiterals) + "}", nil
}

func buildStructExpr(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	sv SnapshotValue,
) (gno.Expr, error) {
	if sv.Kind != string(TypeStruct) {
		return nil, fmt.Errorf("expected struct snapshot value for type %q", typ.String())
	}
	if len(sv.Fields) != len(resolved.Fields) {
		return nil, fmt.Errorf(
			"expected %d struct fields for %q, got %d",
			len(resolved.Fields),
			typ.String(),
			len(sv.Fields),
		)
	}

	structName := typ.Name
	if structName == "" {
		structName = resolved.Name
	}
	if structName == "" {
		return nil, fmt.Errorf("struct literal for %q requires a named struct type", typ.String())
	}

	fields := make(gno.KeyValueExprs, 0, len(resolved.Fields))
	for i, field := range resolved.Fields {
		snapshotField := sv.Fields[i]
		if snapshotField.Name != field.Name {
			return nil, fmt.Errorf(
				"expected struct field %q at index %d for %q, got %q",
				field.Name,
				i,
				typ.String(),
				snapshotField.Name,
			)
		}

		if err := validateStructFieldSnapshotType(schema, field.Type, typ.String(), field.Name); err != nil {
			return nil, err
		}

		expr, err := BuildExpr(schema, field.Type, snapshotField.Value)
		if err != nil {
			return nil, err
		}

		fields = append(fields, gno.Kv(field.Name, expr))
	}

	return &gno.CompositeLitExpr{
		Type: gno.Nx(structName),
		Elts: fields,
	}, nil
}

func extractMapSnapshotValue(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	tv gno.TypedValue,
) (SnapshotValue, error) {
	if resolved.Key == nil || resolved.Elem == nil {
		return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	keyType := schema.ResolveType(*resolved.Key)
	if !keyType.IsScalar() || keyType.NormalizedString() != "string" {
		return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	if !isSupportedSnapshotMapElemType(schema, *resolved.Elem) {
		return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	if tv.V == nil {
		return SnapshotValue{
			Kind:  string(TypeMap),
			IsNil: true,
		}, nil
	}

	mv, ok := tv.V.(*gno.MapValue)
	if !ok || mv == nil {
		return SnapshotValue{}, fmt.Errorf("expected map value for %q", typ.String())
	}

	entries := make([]SnapshotMapEntry, 0)
	if mv.List != nil {
		for item := mv.List.Head; item != nil; item = item.Next {
			key := item.Key.GetString()
			value, err := ExtractSnapshotValue(schema, *resolved.Elem, item.Value)
			if err != nil {
				return SnapshotValue{}, err
			}

			entries = append(entries, SnapshotMapEntry{
				Key:   key,
				Value: value,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	return SnapshotValue{
		Kind:    string(TypeMap),
		Entries: entries,
	}, nil
}

func buildMapLiteral(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	sv SnapshotValue,
) (string, error) {
	if sv.Kind != string(TypeMap) {
		return "", fmt.Errorf("expected map snapshot value for type %q", typ.String())
	}
	if resolved.Key == nil || resolved.Elem == nil {
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	keyType := schema.ResolveType(*resolved.Key)
	if !keyType.IsScalar() || keyType.NormalizedString() != "string" {
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	if !isSupportedSnapshotMapElemType(schema, *resolved.Elem) {
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	if sv.IsNil {
		return "nil", nil
	}

	entries := append([]SnapshotMapEntry(nil), sv.Entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	typeName := typ.String()
	if len(entries) == 0 {
		return typeName + "{}", nil
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		valueLiteral, err := BuildLiteral(schema, *resolved.Elem, entry.Value)
		if err != nil {
			return "", err
		}

		parts = append(parts, strconv.Quote(entry.Key)+":"+valueLiteral)
	}

	return typeName + "{" + joinWithComma(parts) + "}", nil
}

func buildMapExpr(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	sv SnapshotValue,
) (gno.Expr, error) {
	if sv.Kind != string(TypeMap) {
		return nil, fmt.Errorf("expected map snapshot value for type %q", typ.String())
	}
	if resolved.Key == nil || resolved.Elem == nil {
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	keyType := schema.ResolveType(*resolved.Key)
	if !keyType.IsScalar() || keyType.NormalizedString() != "string" {
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	if !isSupportedSnapshotMapElemType(schema, *resolved.Elem) {
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), MapGlobalSupportDescription)
	}

	if sv.IsNil {
		return gno.Nx("nil"), nil
	}

	entries := append([]SnapshotMapEntry(nil), sv.Entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	kvs := make(gno.KeyValueExprs, 0, len(entries))
	for _, entry := range entries {
		valueExpr, err := BuildExpr(schema, *resolved.Elem, entry.Value)
		if err != nil {
			return nil, err
		}

		kvs = append(kvs, gno.Kv(gno.Str(entry.Key), valueExpr))
	}

	return &gno.CompositeLitExpr{
		Type: gno.MapT(buildTypeExpr(*resolved.Key), buildTypeExpr(*resolved.Elem)),
		Elts: kvs,
	}, nil
}

func extractSliceSnapshotValue(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	tv gno.TypedValue,
) (SnapshotValue, error) {
	if resolved.Elem == nil {
		return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), SliceSupportDescription)
	}
	if !isSupportedSnapshotSliceElemType(schema, *resolved.Elem) {
		return SnapshotValue{}, fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), SliceSupportDescription)
	}
	if tv.V == nil {
		return SnapshotValue{
			Kind:  string(TypeSlice),
			IsNil: true,
		}, nil
	}

	sv, ok := tv.V.(*gno.SliceValue)
	if !ok || sv == nil {
		return SnapshotValue{}, fmt.Errorf("expected slice value for %q", typ.String())
	}
	base := sv.GetBase(nil)
	if base == nil {
		return SnapshotValue{}, fmt.Errorf("expected slice base for %q", typ.String())
	}

	items := make([]SnapshotValue, 0, sv.Length)
	for i := 0; i < sv.Length; i++ {
		value, err := ExtractSnapshotValue(schema, *resolved.Elem, base.List[sv.Offset+i])
		if err != nil {
			return SnapshotValue{}, err
		}
		items = append(items, value)
	}

	return SnapshotValue{
		Kind:  string(TypeSlice),
		Items: items,
	}, nil
}

func buildSliceLiteral(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	sv SnapshotValue,
) (string, error) {
	if sv.Kind != string(TypeSlice) {
		return "", fmt.Errorf("expected slice snapshot value for type %q", typ.String())
	}
	if resolved.Elem == nil {
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), SliceSupportDescription)
	}
	if !isSupportedSnapshotSliceElemType(schema, *resolved.Elem) {
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), SliceSupportDescription)
	}
	if sv.IsNil {
		return "nil", nil
	}
	if len(sv.Items) == 0 {
		return typ.String() + "{}", nil
	}

	parts := make([]string, 0, len(sv.Items))
	for _, item := range sv.Items {
		lit, err := BuildLiteral(schema, *resolved.Elem, item)
		if err != nil {
			return "", err
		}
		parts = append(parts, lit)
	}

	return typ.String() + "{" + joinWithComma(parts) + "}", nil
}

func buildSliceExpr(
	schema ContractSchema,
	typ TypeRef,
	resolved TypeRef,
	sv SnapshotValue,
) (gno.Expr, error) {
	if sv.Kind != string(TypeSlice) {
		return nil, fmt.Errorf("expected slice snapshot value for type %q", typ.String())
	}
	if resolved.Elem == nil {
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), SliceSupportDescription)
	}
	if !isSupportedSnapshotSliceElemType(schema, *resolved.Elem) {
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), SliceSupportDescription)
	}
	if sv.IsNil {
		return gno.Nx("nil"), nil
	}

	elts := make(gno.KeyValueExprs, 0, len(sv.Items))
	for _, item := range sv.Items {
		expr, err := BuildExpr(schema, *resolved.Elem, item)
		if err != nil {
			return nil, err
		}
		elts = append(elts, gno.KeyValueExpr{Value: expr})
	}

	return &gno.CompositeLitExpr{
		Type: gno.SliceT(buildTypeExpr(*resolved.Elem)),
		Elts: elts,
	}, nil
}

func isSupportedSnapshotMapElemType(schema ContractSchema, typ TypeRef) bool {
	resolved := schema.ResolveType(typ)
	if resolved.IsScalar() && IsSupportedScalarType(resolved) {
		return true
	}

	_, err := schema.validateNamedStructType(typ, "map value")
	return err == nil
}

func isSupportedSnapshotSliceElemType(schema ContractSchema, typ TypeRef) bool {
	resolved := schema.ResolveType(typ)
	if resolved.IsScalar() && IsSupportedScalarType(resolved) {
		return true
	}

	_, err := schema.validateNamedStructType(typ, "slice element")
	return err == nil
}

func validateStructFieldSnapshotType(schema ContractSchema, fieldType TypeRef, structName string, fieldName string) error {
	resolved := schema.ResolveType(fieldType)

	switch {
	case resolved.IsScalar():
		if IsSupportedScalarType(resolved) {
			return nil
		}
		return fmt.Errorf(
			"struct field %q of %q type %q is not supported",
			fieldName,
			structName,
			fieldType.String(),
		)

	case fieldType.Kind == TypeNamed:
		if _, err := schema.validateNamedStructType(fieldType, fmt.Sprintf("struct field %q of %q", fieldName, structName)); err != nil {
			return err
		}
		return nil

	case resolved.Kind == TypeMap:
		if err := validateMapSnapshotType(schema, fieldType, fmt.Sprintf("struct field %q of %q", fieldName, structName)); err != nil {
			return err
		}
		return nil

	case resolved.Kind == TypeSlice:
		if err := validateSliceSnapshotType(schema, fieldType, fmt.Sprintf("struct field %q of %q", fieldName, structName)); err != nil {
			return err
		}
		return nil

	case fieldType.Kind == TypeStruct:
		return fmt.Errorf(
			"struct field %q of %q type %q is not supported; anonymous nested struct fields are not supported",
			fieldName,
			structName,
			fieldType.String(),
		)

	default:
		return fmt.Errorf(
			"struct field %q of %q type %q is not supported",
			fieldName,
			structName,
			fieldType.String(),
		)
	}
}

func validateSliceSnapshotType(schema ContractSchema, typ TypeRef, subject string) error {
	resolved := schema.ResolveType(typ)
	if resolved.Kind != TypeSlice || resolved.Elem == nil {
		return fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	elemType := schema.ResolveType(*resolved.Elem)
	switch {
	case elemType.IsScalar():
		if IsSupportedScalarType(elemType) {
			return nil
		}
		return fmt.Errorf("%s slice element type %q is not supported", subject, resolved.Elem.String())

	case resolved.Elem.Kind == TypeNamed:
		if _, err := schema.validateNamedStructType(*resolved.Elem, fmt.Sprintf("%s slice element", subject)); err != nil {
			return err
		}
		return nil

	case resolved.Elem.Kind == TypeStruct:
		return fmt.Errorf("%s slice element type %q is not supported; anonymous struct slice elements are not supported", subject, resolved.Elem.String())

	case elemType.Kind == TypeMap:
		return fmt.Errorf("%s slice element type %q is not supported; slice elements cannot be maps", subject, resolved.Elem.String())

	case elemType.Kind == TypeSlice:
		return fmt.Errorf("%s slice element type %q is not supported; slice elements cannot be slices", subject, resolved.Elem.String())

	default:
		return fmt.Errorf("%s slice element type %q is not supported", subject, resolved.Elem.String())
	}
}

func validateMapSnapshotType(schema ContractSchema, typ TypeRef, subject string) error {
	resolved := schema.ResolveType(typ)
	if resolved.Kind != TypeMap || resolved.Key == nil || resolved.Elem == nil {
		return fmt.Errorf("%s type %q is not supported", subject, typ.String())
	}

	keyType := schema.ResolveType(*resolved.Key)
	if !keyType.IsScalar() || keyType.NormalizedString() != "string" {
		return fmt.Errorf("%s map key type %q is not supported", subject, resolved.Key.String())
	}

	elemType := schema.ResolveType(*resolved.Elem)
	switch {
	case elemType.IsScalar():
		if IsSupportedScalarType(elemType) {
			return nil
		}
		return fmt.Errorf("%s map value type %q is not supported", subject, resolved.Elem.String())

	case resolved.Elem.Kind == TypeNamed:
		if _, err := schema.validateNamedStructType(*resolved.Elem, fmt.Sprintf("%s map value", subject)); err != nil {
			return err
		}
		return nil

	case resolved.Elem.Kind == TypeStruct:
		return fmt.Errorf("%s map value type %q is not supported; anonymous struct map values are not supported", subject, resolved.Elem.String())

	case elemType.Kind == TypeMap:
		return fmt.Errorf("%s map value type %q is not supported; map values cannot be maps", subject, resolved.Elem.String())

	case elemType.Kind == TypeSlice:
		return fmt.Errorf("%s map value type %q is not supported; map values cannot be slices", subject, resolved.Elem.String())

	default:
		return fmt.Errorf("%s map value type %q is not supported", subject, resolved.Elem.String())
	}
}

func scalarExpr(typ TypeRef, sv SnapshotValue) (gno.Expr, error) {
	if sv.Kind != string(TypeScalar) {
		return nil, fmt.Errorf("expected scalar snapshot value for type %q", typ.String())
	}

	switch typ.NormalizedString() {
	case "string":
		return gno.Str(sv.Scalar), nil
	case "bool":
		if sv.Scalar != "true" && sv.Scalar != "false" {
			return nil, fmt.Errorf("invalid bool literal %q", sv.Scalar)
		}
		return gno.X(sv.Scalar), nil
	case "int", "int64", "uint64":
		return gno.Num(sv.Scalar), nil
	default:
		return nil, fmt.Errorf("unsupported restore type %q; %s", typ.String(), ScalarOnlySupportDescription)
	}
}

func buildTypeExpr(typ TypeRef) gno.Expr {
	switch typ.Kind {
	case TypeScalar, TypeNamed:
		return gno.Nx(typ.String())
	case TypeMap:
		return gno.MapT(buildTypeExpr(*typ.Key), buildTypeExpr(*typ.Elem))
	case TypeSlice:
		return gno.SliceT(buildTypeExpr(*typ.Elem))
	default:
		return gno.Nx(typ.String())
	}
}

func joinWithComma(items []string) string {
	if len(items) == 0 {
		return ""
	}

	out := items[0]
	for i := 1; i < len(items); i++ {
		out += "," + items[i]
	}

	return out
}

func scalarTypedValueToString(typ TypeRef, tv gno.TypedValue) (string, error) {
	switch typ.NormalizedString() {
	case "string":
		return tv.GetString(), nil
	case "bool":
		if tv.GetBool() {
			return "true", nil
		}
		return "false", nil
	case "int":
		return strconv.FormatInt(tv.GetInt(), 10), nil
	case "int64":
		return strconv.FormatInt(tv.GetInt64(), 10), nil
	case "uint64":
		return strconv.FormatUint(tv.GetUint64(), 10), nil
	default:
		return "", fmt.Errorf("unsupported snapshot type %q; %s", typ.String(), ScalarOnlySupportDescription)
	}
}

func scalarLiteral(typ TypeRef, sv SnapshotValue) (string, error) {
	if sv.Kind != string(TypeScalar) {
		return "", fmt.Errorf("expected scalar snapshot value for type %q", typ.String())
	}

	switch typ.NormalizedString() {
	case "string":
		return strconv.Quote(sv.Scalar), nil
	case "bool", "int", "int64", "uint64":
		return sv.Scalar, nil
	default:
		return "", fmt.Errorf("unsupported restore type %q; %s", typ.String(), ScalarOnlySupportDescription)
	}
}
