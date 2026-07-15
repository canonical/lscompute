// Package visualization holds presentation-only copies of the core machine
// data model. It decouples human-readable rendering (custom YAML formatting)
// from pkg/machine, so the core types can stay pure data. Convert a
// machine.MachineInfo with New before marshalling it for display.
package visualization

import (
	"github.com/canonical/lscompute/pkg/machine"
	"github.com/canonical/lscompute/pkg/machine/types"

	"go.yaml.in/yaml/v4"
)

// MachineInfo is the visualization copy of machine.MachineInfo. It mirrors the
// core struct but uses visualization sub-types that carry the custom YAML
// rendering.
type MachineInfo struct {
	Cpus    []CpuInfo          `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  MemoryInfo         `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    map[string]DirInfo `json:"disk,omitempty" yaml:"disk,omitempty"`
	Devices []any              `json:"devices,omitempty" yaml:"devices,omitempty"`
}

// CpuInfo mirrors cpu.CpuInfo but uses the visualization FlagList so CPU flags
// render as an inline (flow-style) YAML sequence.
type CpuInfo struct {
	Architecture string `json:"architecture" yaml:"architecture"`

	// amd64
	ManufacturerId string   `json:"manufacturer-id,omitempty" yaml:"manufacturer-id,omitempty"`
	Flags          FlagList `json:"flags,omitempty" yaml:"flags,omitempty"`

	// arm64
	ImplementerId types.HexInt `json:"implementer-id,omitempty" yaml:"implementer-id,omitempty"`
	PartNumber    types.HexInt `json:"part-number,omitempty" yaml:"part-number,omitempty"`
	Features      []string     `json:"features,omitempty" yaml:"features,omitempty"`

	// riscv64
	Isa []string `json:"isa,omitempty" yaml:"isa,omitempty"`
}

// MemoryInfo mirrors memory.MemoryInfo and renders capacities using
// human-readable byte magnitudes in YAML while leaving JSON as raw byte counts.
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

// DirInfo mirrors disk.DirInfo and renders capacities using human-readable byte
// magnitudes in YAML while leaving JSON as raw byte counts.
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

// FlagList is a list of CPU flags rendered as an inline (flow-style) YAML
// sequence, e.g. [fpu, vme, de]. JSON output is unaffected.
type FlagList []string

// MarshalYAML renders the flag list as a flow-style sequence.
func (f FlagList) MarshalYAML() (any, error) {
	node := &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
	for _, flag := range f {
		node.Content = append(node.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: flag,
		})
	}
	return node, nil
}

// New builds a visualization MachineInfo from a core machine.MachineInfo. It
// returns nil if info is nil. Devices are carried over unchanged; their custom
// rendering is provided by the leaf types (for example types.HexInt).
func New(info *machine.MachineInfo) *MachineInfo {
	if info == nil {
		return nil
	}

	v := &MachineInfo{
		Memory:  MemoryInfo(info.Memory),
		Devices: info.Devices,
	}

	if info.Cpus != nil {
		v.Cpus = make([]CpuInfo, len(info.Cpus))
		for i, c := range info.Cpus {
			v.Cpus[i] = CpuInfo{
				Architecture:   c.Architecture,
				ManufacturerId: c.ManufacturerId,
				Flags:          FlagList(c.Flags),
				ImplementerId:  c.ImplementerId,
				PartNumber:     c.PartNumber,
				Features:       c.Features,
				Isa:            c.Isa,
			}
		}
	}

	if info.Disk != nil {
		v.Disk = make(map[string]DirInfo, len(info.Disk))
		for path, d := range info.Disk {
			v.Disk[path] = DirInfo(d)
		}
	}

	return v
}
