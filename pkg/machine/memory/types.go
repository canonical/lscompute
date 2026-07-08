package memory

import "github.com/canonical/lscompute/pkg/machine/types"

type MemoryInfo struct {
	TotalRam  uint64 `json:"total-ram" yaml:"total-ram"`
	TotalSwap uint64 `json:"total-swap" yaml:"total-swap"`
}

// MarshalYAML renders memory capacities using human-readable byte magnitudes
// (M, G, T suffixes) while leaving JSON output as raw byte counts.
func (m MemoryInfo) MarshalYAML() (any, error) {
	return struct {
		TotalRam  any `yaml:"total-ram"`
		TotalSwap any `yaml:"total-swap"`
	}{
		TotalRam:  types.FormatBytes(m.TotalRam),
		TotalSwap: types.FormatBytes(m.TotalSwap),
	}, nil
}
