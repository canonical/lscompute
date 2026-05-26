package machine

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/memory"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func Get(h host.Host, friendlyNames bool) (*types.MachineInfo, []string, error) {
	var machineInfo types.MachineInfo

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

	devices, warnings, err := Devices(h, friendlyNames)
	if err != nil {
		return nil, nil, fmt.Errorf("getting devices: %w", err)
	}
	machineInfo.Devices = devices

	return &machineInfo, warnings, nil
}
