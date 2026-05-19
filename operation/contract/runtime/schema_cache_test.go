package runtime

import "sync"

func resetContractSchemaCacheForTest() {
	contractSchemaCache = sync.Map{}
}
