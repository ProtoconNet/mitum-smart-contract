package runtime

import "fmt"

const (
	MaxContractCallDataEntries    = 64
	MaxContractCallDataKeyBytes   = 128
	MaxContractCallDataValueBytes = 16 * 1024
	MaxContractCallDataTotalBytes = 64 * 1024
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
