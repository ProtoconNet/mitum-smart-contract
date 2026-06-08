package runtime

import "fmt"

const (
	MaxContractCallDataEntries     = 64
	MaxContractCallDataKeyBytes    = 128
	MaxContractCallDataValueBytes  = 16 * 1024
	MaxContractCallDataTotalBytes  = 64 * 1024
	MaxContractCallItems           = 16
	MaxContractCallItemsTotalBytes = 64 * 1024
)

func ValidateContractCallDataLimits(name string, callData map[string]string) error {
	if name == "" {
		name = "contract call data"
	}

	if len(callData) > MaxContractCallDataEntries {
		return fmt.Errorf(
			"%s exceeds max entries: got %d, max %d",
			name,
			len(callData),
			MaxContractCallDataEntries,
		)
	}

	total := 0
	for key, value := range callData {
		keyBytes := len(key)
		if keyBytes > MaxContractCallDataKeyBytes {
			return fmt.Errorf(
				"%s key exceeds max size: got %d bytes, max %d bytes",
				name,
				keyBytes,
				MaxContractCallDataKeyBytes,
			)
		}

		valueBytes := len(value)
		if valueBytes > MaxContractCallDataValueBytes {
			return fmt.Errorf(
				"%s value for key %q exceeds max size: got %d bytes, max %d bytes",
				name,
				key,
				valueBytes,
				MaxContractCallDataValueBytes,
			)
		}

		total += keyBytes + valueBytes
		if total > MaxContractCallDataTotalBytes {
			return fmt.Errorf(
				"%s exceeds max total key+value size: got %d bytes, max %d bytes",
				name,
				total,
				MaxContractCallDataTotalBytes,
			)
		}
	}

	return nil
}

func ValidateContractCallItemsLimits(name string, items []ExecuteCallItem) error {
	if name == "" {
		name = "contract call items"
	}

	if len(items) < 1 {
		return fmt.Errorf("%s is empty", name)
	}
	if len(items) > MaxContractCallItems {
		return fmt.Errorf(
			"%s exceeds max items: got %d, max %d",
			name,
			len(items),
			MaxContractCallItems,
		)
	}

	total := 0
	for i := range items {
		item := items[i]
		if item.Function == "" {
			return fmt.Errorf("%s item %d has empty function", name, i+1)
		}
		if _, found := item.CallData["function"]; found {
			return fmt.Errorf("%s item %d call_data must not include function selector key", name, i+1)
		}
		if err := ValidateContractCallDataLimits(
			fmt.Sprintf("%s item %d call_data", name, i+1),
			item.CallData,
		); err != nil {
			return err
		}

		total += len(item.Function) + ContractCallDataTotalBytes(item.CallData)
		if total > MaxContractCallItemsTotalBytes {
			return fmt.Errorf(
				"%s exceeds max total function+call_data size: got %d bytes, max %d bytes",
				name,
				total,
				MaxContractCallItemsTotalBytes,
			)
		}
	}

	return nil
}

func ContractCallDataTotalBytes(callData map[string]string) int {
	total := 0
	for key, value := range callData {
		total += len(key) + len(value)
	}

	return total
}
