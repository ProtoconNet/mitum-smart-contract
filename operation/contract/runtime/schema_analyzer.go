package runtime

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"github.com/pkg/errors"
)

type typeResolver struct {
	fset      *token.FileSet
	typeDecls map[string]ast.Expr
	resolved  map[string]TypeRef
	resolving map[string]bool
}

func newTypeResolver(fset *token.FileSet, node *ast.File) *typeResolver {
	typeDecls := map[string]ast.Expr{}

	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typeDecls[ts.Name.Name] = ts.Type
		}
	}

	return &typeResolver{
		fset:      fset,
		typeDecls: typeDecls,
		resolved:  map[string]TypeRef{},
		resolving: map[string]bool{},
	}
}

func AnalyzeContractSchema(sourceCode string) (ContractSchema, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", sourceCode, parser.AllErrors)
	if err != nil {
		return ContractSchema{}, errors.Wrap(err, "failed to parse contract source for schema analysis")
	}

	resolver := newTypeResolver(fset, node)
	schema := ContractSchema{
		PackageName: node.Name.Name,
		Types:       NewTypeRegistry(),
	}

	if err := schema.Types.populateFromResolver(resolver); err != nil {
		return ContractSchema{}, err
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.VAR {
				continue
			}

			globals, err := parsePersistentBindings(resolver, d)
			if err != nil {
				return ContractSchema{}, err
			}

			schema.PersistentGlobals = append(schema.PersistentGlobals, globals...)

		case *ast.FuncDecl:
			if d.Recv != nil {
				continue
			}

			fn, err := parseFunctionSchema(resolver, d)
			if err != nil {
				return ContractSchema{}, err
			}

			schema.Functions = append(schema.Functions, fn)
		}
	}

	if err := finalizeContractSchema(&schema); err != nil {
		return ContractSchema{}, err
	}

	return schema, nil
}

func (r *TypeRegistry) populateFromResolver(resolver *typeResolver) error {
	if r.Structs == nil {
		r.Structs = map[string]TypeRef{}
	}

	for name := range resolver.typeDecls {
		tref, err := resolver.resolveNamedType(name)
		if err != nil {
			return err
		}

		if tref.Kind == TypeStruct {
			r.Structs[name] = tref
		}
	}

	return nil
}

func parsePersistentBindings(
	resolver *typeResolver,
	decl *ast.GenDecl,
) ([]PersistentBindingSchema, error) {
	var bindings []PersistentBindingSchema

	for _, spec := range decl.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		typ := TypeRef{}
		if vs.Type != nil {
			var err error
			typ, err = resolver.resolveTypeExpr(vs.Type)
			if err != nil {
				return nil, errors.Wrap(err, "failed to resolve persistent binding type")
			}
		}

		for _, name := range vs.Names {
			bindings = append(bindings, PersistentBindingSchema{
				Name:            name.Name,
				Type:            typ,
				HasExplicitType: vs.Type != nil,
			})
		}
	}

	return bindings, nil
}

func parseFunctionSchema(
	resolver *typeResolver,
	fn *ast.FuncDecl,
) (FunctionSchema, error) {
	out := FunctionSchema{
		Name:     fn.Name.Name,
		Exported: ast.IsExported(fn.Name.Name),
	}

	params, err := parseFieldList(resolver, fn.Type.Params)
	if err != nil {
		return FunctionSchema{}, errors.Wrapf(err, "failed to parse params of %s", fn.Name.Name)
	}
	out.Params = params

	results, err := parseResultList(resolver, fn.Type.Results)
	if err != nil {
		return FunctionSchema{}, errors.Wrapf(err, "failed to parse results of %s", fn.Name.Name)
	}
	out.Results = results

	return out, nil
}

func parseFieldList(
	resolver *typeResolver,
	fields *ast.FieldList,
) ([]ParamSchema, error) {
	if fields == nil {
		return nil, nil
	}

	var out []ParamSchema

	for _, field := range fields.List {
		typ, err := resolver.resolveTypeExpr(field.Type)
		if err != nil {
			return nil, err
		}

		if len(field.Names) == 0 {
			out = append(out, ParamSchema{
				Name: "",
				Type: typ,
			})
			continue
		}

		for _, name := range field.Names {
			out = append(out, ParamSchema{
				Name: name.Name,
				Type: typ,
			})
		}
	}

	return out, nil
}

func parseResultList(
	resolver *typeResolver,
	fields *ast.FieldList,
) ([]ResultSchema, error) {
	if fields == nil {
		return nil, nil
	}

	var out []ResultSchema

	for _, field := range fields.List {
		typ, err := resolver.resolveTypeExpr(field.Type)
		if err != nil {
			return nil, err
		}

		count := len(field.Names)
		if count == 0 {
			count = 1
		}

		for i := 0; i < count; i++ {
			out = append(out, ResultSchema{
				Type: typ,
			})
		}
	}

	return out, nil
}

func (r *typeResolver) resolveNamedType(name string) (TypeRef, error) {
	if tref, found := r.resolved[name]; found {
		return tref, nil
	}
	if r.resolving[name] {
		return TypeRef{}, errors.Errorf("recursive type declaration %q is not supported yet", name)
	}

	expr, found := r.typeDecls[name]
	if !found {
		return TypeRef{}, errors.Errorf("type declaration %q not found", name)
	}

	r.resolving[name] = true
	defer delete(r.resolving, name)

	tref, err := r.resolveTypeExpr(expr)
	if err != nil {
		return TypeRef{}, err
	}

	if tref.Name == "" {
		tref.Name = name
	}
	if tref.Raw == "" {
		tref.Raw = name
	}

	r.resolved[name] = tref
	return tref, nil
}

func (r *typeResolver) resolveTypeExpr(expr ast.Expr) (TypeRef, error) {
	raw, err := exprString(r.fset, expr)
	if err != nil {
		return TypeRef{}, err
	}

	switch e := expr.(type) {
	case *ast.Ident:
		switch normalizeTypeString(e.Name) {
		case "string", "bool", "int", "int64", "uint64":
			return TypeRef{
				Kind:   TypeScalar,
				Raw:    raw,
				Scalar: normalizeTypeString(e.Name),
			}, nil
		default:
			if _, found := r.typeDecls[e.Name]; found {
				return TypeRef{
					Kind: TypeNamed,
					Raw:  raw,
					Name: e.Name,
				}, nil
			}

			return TypeRef{
				Kind: TypeOpaque,
				Raw:  raw,
				Name: e.Name,
			}, nil
		}

	case *ast.SelectorExpr:
		return TypeRef{
			Kind: TypeOpaque,
			Raw:  raw,
			Name: raw,
		}, nil

	case *ast.InterfaceType:
		return TypeRef{
			Kind: TypeOpaque,
			Raw:  raw,
			Name: raw,
		}, nil

	case *ast.MapType:
		key, err := r.resolveTypeExpr(e.Key)
		if err != nil {
			return TypeRef{}, err
		}
		elem, err := r.resolveTypeExpr(e.Value)
		if err != nil {
			return TypeRef{}, err
		}
		return TypeRef{
			Kind: TypeMap,
			Raw:  raw,
			Key:  &key,
			Elem: &elem,
		}, nil

	case *ast.ArrayType:
		if e.Len != nil {
			return TypeRef{
				Kind: TypeOpaque,
				Raw:  raw,
				Name: raw,
			}, nil
		}

		elem, err := r.resolveTypeExpr(e.Elt)
		if err != nil {
			return TypeRef{}, err
		}
		return TypeRef{
			Kind: TypeSlice,
			Raw:  raw,
			Elem: &elem,
		}, nil

	case *ast.StructType:
		fields, err := r.resolveStructFields(e.Fields)
		if err != nil {
			return TypeRef{}, err
		}
		return TypeRef{
			Kind:   TypeStruct,
			Raw:    raw,
			Fields: fields,
		}, nil

	default:
		return TypeRef{
			Kind: TypeOpaque,
			Raw:  raw,
			Name: raw,
		}, nil
	}
}

func (r *typeResolver) resolveStructFields(fields *ast.FieldList) ([]StructField, error) {
	if fields == nil {
		return nil, nil
	}

	var out []StructField

	for _, field := range fields.List {
		typ, err := r.resolveTypeExpr(field.Type)
		if err != nil {
			return nil, err
		}

		if len(field.Names) == 0 {
			return nil, errors.Errorf("embedded struct fields are not supported yet")
		}

		for _, name := range field.Names {
			out = append(out, StructField{
				Name: name.Name,
				Type: typ,
			})
		}
	}

	return out, nil
}

func exprString(fset *token.FileSet, expr ast.Expr) (string, error) {
	var buf bytes.Buffer

	if err := format.Node(&buf, fset, expr); err != nil {
		return "", err
	}

	return normalizeTypeString(buf.String()), nil
}

func finalizeContractSchema(schema *ContractSchema) error {
	initialize, found := schema.FindFunction("Initialize")
	if !found {
		return errors.Errorf("Initialize function not found")
	}

	if schema.PackageName != "contract" {
		return errors.Errorf("only package contract typed Gno contracts are supported")
	}

	if !initialize.IsTypedInitializeShape() {
		return errors.Errorf("typed Gno contract Initialize must be func Initialize(ctx ContractContext) error")
	}

	for _, g := range schema.PersistentGlobals {
		if !g.HasExplicitType {
			return errors.Errorf("persistent global %q must declare explicit type", g.Name)
		}
		if err := schema.ValidatePersistentGlobalType(g); err != nil {
			return err
		}
	}

	for _, fn := range schema.Functions {
		if fn.Name == "Initialize" {
			continue
		}
		if !fn.Exported || !fn.IsContextCallable() {
			continue
		}

		if fn.IsTypedWriteShape() {
			for i := 1; i < len(fn.Params); i++ {
				if err := schema.ValidateWriteArgType(fn.Params[i].Type, fmt.Sprintf("function %q parameter %q", fn.Name, fn.Params[i].Name)); err != nil {
					return errors.Wrap(err, "invalid write parameter type")
				}
			}
			continue
		}

		if fn.IsTypedQueryShape() {
			for i := 1; i < len(fn.Params); i++ {
				if err := schema.ValidateQueryArgType(fn.Params[i].Type, fmt.Sprintf("query function %q parameter %q", fn.Name, fn.Params[i].Name)); err != nil {
					return errors.Wrap(err, "invalid query parameter type")
				}
			}

			if err := schema.ValidateQueryResultType(fn.Results[0].Type, fmt.Sprintf("query function %q result", fn.Name)); err != nil {
				return errors.Wrap(err, "invalid query result type")
			}
			if len(fn.Results) == 2 && fn.Results[1].Type.NormalizedString() != "bool" {
				return errors.Errorf("query function %q second result must be bool", fn.Name)
			}

			continue
		}

		return errors.Errorf("typed contract function %q must be either write(ctx, ...) error or query(ctx, ...) T[/bool]", fn.Name)
	}

	schema.Mode = SchemaModeTypedArgs

	return nil
}
