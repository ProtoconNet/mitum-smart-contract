package runtime

import (
	"crypto/sha256"
	"sync"
)

var contractSchemaCache sync.Map

func contractSchemaCacheKey(sourceCode string) [sha256.Size]byte {
	return sha256.Sum256([]byte(sourceCode))
}

func loadContractSchemaFromCache(sourceCode string) (ContractSchema, bool) {
	cached, found := contractSchemaCache.Load(contractSchemaCacheKey(sourceCode))
	if !found {
		return ContractSchema{}, false
	}

	schema, ok := cached.(ContractSchema)
	if !ok {
		return ContractSchema{}, false
	}

	return schema, true
}

func storeContractSchemaInCache(sourceCode string, schema ContractSchema) {
	contractSchemaCache.Store(contractSchemaCacheKey(sourceCode), schema)
}
