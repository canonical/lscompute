package disk

import "github.com/canonical/lscompute/pkg/machine/types"

type DirInfo struct {
	Total uint64 `json:"total" yaml:"total"`
	Avail uint64 `json:"avail" yaml:"avail"`
}

// MarshalYAML renders disk capacities using human-readable byte magnitudes
// (M, G, T suffixes) while leaving JSON output as raw byte counts.
func (d DirInfo) MarshalYAML() (any, error) {
	return struct {
		Total any `yaml:"total"`
		Avail any `yaml:"avail"`
	}{
		Total: types.FormatBytes(d.Total),
		Avail: types.FormatBytes(d.Avail),
	}, nil
}
