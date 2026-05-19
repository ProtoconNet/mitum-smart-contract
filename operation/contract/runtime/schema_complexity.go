package runtime

import "fmt"

const (
	MaxContractSchemaImports           = 16
	MaxContractSchemaFunctions         = 128
	MaxContractSchemaPersistentGlobals = 128
	MaxContractSchemaStructs           = 64
	MaxContractSchemaStructFields      = 64
	MaxContractSchemaTypeNestingDepth  = 16
	MaxContractSchemaNodes             = 4096
)

func validateContractSchemaComplexity(schema ContractSchema, importCount int) error {
	if importCount > MaxContractSchemaImports {
		return fmt.Errorf(
			"contract schema complexity exceeds max import count: got %d, max %d",
			importCount,
			MaxContractSchemaImports,
		)
	}

	if len(schema.Functions) > MaxContractSchemaFunctions {
		return fmt.Errorf(
			"contract schema complexity exceeds max function count: got %d, max %d",
			len(schema.Functions),
			MaxContractSchemaFunctions,
		)
	}

	if len(schema.PersistentGlobals) > MaxContractSchemaPersistentGlobals {
		return fmt.Errorf(
			"contract schema complexity exceeds max persistent global count: got %d, max %d",
			len(schema.PersistentGlobals),
			MaxContractSchemaPersistentGlobals,
		)
	}

	if len(schema.Types.Structs) > MaxContractSchemaStructs {
		return fmt.Errorf(
			"contract schema complexity exceeds max struct count: got %d, max %d",
			len(schema.Types.Structs),
			MaxContractSchemaStructs,
		)
	}

	for name, typ := range schema.Types.Structs {
		if len(typ.Fields) > MaxContractSchemaStructFields {
			return fmt.Errorf(
				"contract schema complexity exceeds max struct field count for %q: got %d, max %d",
				name,
				len(typ.Fields),
				MaxContractSchemaStructFields,
			)
		}
	}

	if depth := maxContractSchemaTypeDepth(schema); depth > MaxContractSchemaTypeNestingDepth {
		return fmt.Errorf(
			"contract schema complexity exceeds max type nesting depth: got %d, max %d",
			depth,
			MaxContractSchemaTypeNestingDepth,
		)
	}

	nodes := contractSchemaNodeCount(schema, importCount)
	if nodes > MaxContractSchemaNodes {
		return fmt.Errorf(
			"contract schema complexity exceeds max total node count: got %d, max %d",
			nodes,
			MaxContractSchemaNodes,
		)
	}

	return nil
}

func maxContractSchemaTypeDepth(schema ContractSchema) int {
	maxDepth := 0

	for _, typ := range schema.Types.Structs {
		maxDepth = max(maxDepth, contractSchemaTypeDepth(schema, typ, map[string]bool{}))
	}
	for _, global := range schema.PersistentGlobals {
		maxDepth = max(maxDepth, contractSchemaTypeDepth(schema, global.Type, map[string]bool{}))
	}
	for _, fn := range schema.Functions {
		for _, param := range fn.Params {
			maxDepth = max(maxDepth, contractSchemaTypeDepth(schema, param.Type, map[string]bool{}))
		}
		for _, result := range fn.Results {
			maxDepth = max(maxDepth, contractSchemaTypeDepth(schema, result.Type, map[string]bool{}))
		}
	}

	return maxDepth
}

func contractSchemaTypeDepth(schema ContractSchema, typ TypeRef, resolving map[string]bool) int {
	switch typ.Kind {
	case TypeNamed:
		if resolving[typ.Name] {
			return 0
		}

		resolved, found := schema.Types.Resolve(typ)
		if !found {
			return 0
		}

		return contractSchemaTypeDepth(schema, resolved, resolving)

	case TypeStruct:
		name := typ.Name
		if name != "" {
			if resolving[name] {
				return 0
			}
			resolving[name] = true
			defer delete(resolving, name)
		}

		maxFieldDepth := 0
		for _, field := range typ.Fields {
			maxFieldDepth = max(maxFieldDepth, contractSchemaTypeDepth(schema, field.Type, resolving))
		}

		return 1 + maxFieldDepth

	case TypeMap:
		keyDepth := 0
		if typ.Key != nil {
			keyDepth = contractSchemaTypeDepth(schema, *typ.Key, resolving)
		}
		elemDepth := 0
		if typ.Elem != nil {
			elemDepth = contractSchemaTypeDepth(schema, *typ.Elem, resolving)
		}

		return 1 + max(keyDepth, elemDepth)

	case TypeSlice:
		if typ.Elem == nil {
			return 1
		}

		return 1 + contractSchemaTypeDepth(schema, *typ.Elem, resolving)

	default:
		return 0
	}
}

func contractSchemaNodeCount(schema ContractSchema, importCount int) int {
	nodes := importCount +
		len(schema.Functions) +
		len(schema.PersistentGlobals) +
		len(schema.Types.Structs)

	for _, typ := range schema.Types.Structs {
		nodes += len(typ.Fields)
		for _, field := range typ.Fields {
			nodes += contractSchemaTypeNodeCount(schema, field.Type, map[string]bool{})
		}
	}
	for _, global := range schema.PersistentGlobals {
		nodes += contractSchemaTypeNodeCount(schema, global.Type, map[string]bool{})
	}
	for _, fn := range schema.Functions {
		nodes += len(fn.Params) + len(fn.Results)
		for _, param := range fn.Params {
			nodes += contractSchemaTypeNodeCount(schema, param.Type, map[string]bool{})
		}
		for _, result := range fn.Results {
			nodes += contractSchemaTypeNodeCount(schema, result.Type, map[string]bool{})
		}
	}

	return nodes
}

func contractSchemaTypeNodeCount(schema ContractSchema, typ TypeRef, resolving map[string]bool) int {
	switch typ.Kind {
	case TypeNamed:
		if resolving[typ.Name] {
			return 1
		}

		resolved, found := schema.Types.Resolve(typ)
		if !found {
			return 1
		}

		return 1 + contractSchemaTypeNodeCount(schema, resolved, resolving)

	case TypeStruct:
		name := typ.Name
		if name != "" {
			if resolving[name] {
				return 1
			}
			resolving[name] = true
			defer delete(resolving, name)
		}

		nodes := 1 + len(typ.Fields)
		for _, field := range typ.Fields {
			nodes += contractSchemaTypeNodeCount(schema, field.Type, resolving)
		}

		return nodes

	case TypeMap:
		nodes := 1
		if typ.Key != nil {
			nodes += contractSchemaTypeNodeCount(schema, *typ.Key, resolving)
		}
		if typ.Elem != nil {
			nodes += contractSchemaTypeNodeCount(schema, *typ.Elem, resolving)
		}

		return nodes

	case TypeSlice:
		if typ.Elem == nil {
			return 1
		}

		return 1 + contractSchemaTypeNodeCount(schema, *typ.Elem, resolving)

	default:
		return 1
	}
}
