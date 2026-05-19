package runtime

import (
	"fmt"
	"strconv"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func buildInitializeCallArgExpr(schema ContractSchema, param ParamSchema, raw string) (gno.Expr, error) {
	resolved := schema.ResolveType(param.Type)
	if !resolved.IsScalar() || !IsSupportedScalarType(resolved) {
		return nil, fmt.Errorf("unsupported initialize arg type %q; %s", param.Type.String(), WriteArgSupportDescription)
	}

	arg, err := scalarArgExpr(resolved, raw, WriteArgSupportDescription)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse initialize arg %q as %s: %w`, param.Name, resolved.NormalizedString(), err)
	}

	return arg, nil
}

func buildWriteCallArgExpr(schema ContractSchema, typ TypeRef, raw string) (gno.Expr, error) {
	return buildScalarCallArgExpr(schema, typ, raw, WriteArgSupportDescription)
}

func buildQueryCallArgExpr(schema ContractSchema, typ TypeRef, raw string) (gno.Expr, error) {
	return buildScalarCallArgExpr(schema, typ, raw, QueryArgSupportDescription)
}

func buildScalarCallArgExpr(schema ContractSchema, typ TypeRef, raw string, description string) (gno.Expr, error) {
	resolved := schema.ResolveType(typ)
	if !resolved.IsScalar() || !IsSupportedScalarType(resolved) {
		return nil, fmt.Errorf("unsupported arg type %q; %s", typ.String(), description)
	}

	return scalarArgExpr(resolved, raw, description)
}

func scalarArgExpr(typ TypeRef, raw string, description string) (gno.Expr, error) {
	switch typ.NormalizedString() {
	case "string":
		return gno.Str(raw), nil
	case "bool":
		if raw != "true" && raw != "false" {
			return nil, fmt.Errorf("expected bool string")
		}
		return gno.X(raw), nil
	case "int", "int64":
		if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
			return nil, err
		}
		return gno.Num(raw), nil
	case "uint64":
		if _, err := strconv.ParseUint(raw, 10, 64); err != nil {
			return nil, err
		}
		return gno.Num(raw), nil
	default:
		return nil, fmt.Errorf("unsupported arg type %q; %s", typ.String(), description)
	}
}
