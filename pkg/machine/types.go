package machine

import (
	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

type MachineInfo struct {
	Cpus    []cpu.CpuInfo            `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  memory.MemoryInfo        `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    map[string]disk.DirStats `json:"disk,omitempty" yaml:"disk,omitempty"`
	Devices []bus.DeviceInfo         `json:"devices,omitempty" yaml:"devices,omitempty"`
}
