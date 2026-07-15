package machine_visualization

import (
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// VisualizationMachineInfo is a copy of machine.MachineInfo used to decouple
// the visualization layer from the core data model. It mirrors the fields of
// machine.MachineInfo so that presentation concerns can evolve independently
// of the underlying collection logic.
type VisualizationMachineInfo struct {
	Cpus    []cpu.CpuInfo           `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  memory.MemoryInfo       `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    map[string]disk.DirInfo `json:"disk,omitempty" yaml:"disk,omitempty"`
	Devices []any                   `json:"devices,omitempty" yaml:"devices,omitempty"`
}
