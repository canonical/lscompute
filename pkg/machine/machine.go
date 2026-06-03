package machine

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

type MachineInfo struct {
	Cpus    []cpu.CpuInfo           `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Memory  memory.MemoryInfo       `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk    map[string]disk.DirInfo `json:"disk,omitempty" yaml:"disk,omitempty"`
	Devices []any                   `json:"devices,omitempty" yaml:"devices,omitempty"`
}

func Get(h host.Host, friendlyNames bool) (*MachineInfo, []string, error) {
	var machineInfo MachineInfo

	memoryInfo, err := memory.Info(h)
	if err != nil {
		return nil, nil, fmt.Errorf("getting memory info: %w", err)
	}
	machineInfo.Memory = memoryInfo

	cpus, err := cpu.Info(h)
	if err != nil {
		return nil, nil, fmt.Errorf("getting cpu info: %w", err)
	}
	machineInfo.Cpus = cpus

	diskInfo, err := disk.Info(h)
	if err != nil {
		return nil, nil, fmt.Errorf("getting disk info: %w", err)
	}
	machineInfo.Disk = diskInfo

	devices, warnings, err := device.Devices(h, friendlyNames)
	if err != nil {
		return nil, nil, fmt.Errorf("getting devices: %w", err)
	}
	machineInfo.Devices = devices

	return &machineInfo, warnings, nil
}
