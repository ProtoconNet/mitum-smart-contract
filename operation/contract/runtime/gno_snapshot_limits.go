package runtime

import "fmt"

const (
	MaxContractSnapshotBytes      = 256 * 1024
	MaxContractSnapshotMapEntries = 256
	MaxContractSnapshotSliceItems = 256
	MaxContractSnapshotNodes      = 8 * 1024
)

type SnapshotStats struct {
	Bytes       int
	Bindings    int
	MapEntries  int
	SliceItems  int
	StructNodes int
	Nodes       int
}

func SnapshotStatsForDoc(doc SnapshotDoc, snapshotBytes []byte) SnapshotStats {
	stats := SnapshotStats{
		Bytes:    len(snapshotBytes),
		Bindings: len(doc.Bindings),
	}

	for _, binding := range doc.Bindings {
		accumulateSnapshotValueStats(binding.Value, &stats)
	}

	return stats
}

func ValidateSnapshotLimits(doc SnapshotDoc, snapshotBytes []byte) error {
	stats := SnapshotStatsForDoc(doc, snapshotBytes)
	return validateSnapshotStats(stats)
}

func validateSnapshotSizeLimit(snapshotBytes []byte) error {
	if len(snapshotBytes) > MaxContractSnapshotBytes {
		return fmt.Errorf(
			"snapshot exceeds max size: %d > %d bytes",
			len(snapshotBytes),
			MaxContractSnapshotBytes,
		)
	}

	return nil
}

func validateSnapshotStats(stats SnapshotStats) error {
	if stats.Bytes > MaxContractSnapshotBytes {
		return fmt.Errorf(
			"snapshot exceeds max size: %d > %d bytes",
			stats.Bytes,
			MaxContractSnapshotBytes,
		)
	}
	if stats.MapEntries > MaxContractSnapshotMapEntries {
		return fmt.Errorf(
			"snapshot exceeds max map entries: %d > %d",
			stats.MapEntries,
			MaxContractSnapshotMapEntries,
		)
	}
	if stats.SliceItems > MaxContractSnapshotSliceItems {
		return fmt.Errorf(
			"snapshot exceeds max slice items: %d > %d",
			stats.SliceItems,
			MaxContractSnapshotSliceItems,
		)
	}
	if stats.Nodes > MaxContractSnapshotNodes {
		return fmt.Errorf(
			"snapshot exceeds max nodes: %d > %d",
			stats.Nodes,
			MaxContractSnapshotNodes,
		)
	}

	return nil
}

func accumulateSnapshotValueStats(value SnapshotValue, stats *SnapshotStats) {
	stats.Nodes++

	switch value.Kind {
	case string(TypeStruct):
		stats.StructNodes++
		for _, field := range value.Fields {
			accumulateSnapshotValueStats(field.Value, stats)
		}
	case string(TypeMap):
		stats.MapEntries += len(value.Entries)
		for _, entry := range value.Entries {
			accumulateSnapshotValueStats(entry.Value, stats)
		}
	case string(TypeSlice):
		stats.SliceItems += len(value.Items)
		for _, item := range value.Items {
			accumulateSnapshotValueStats(item, stats)
		}
	}
}
