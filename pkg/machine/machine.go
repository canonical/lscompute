package machine

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

type Machine struct {
	CPUs           []cpu.CPU
	Memory         memory.Memory
	Disk           []disk.Disk
	PCIDevices     []pci.Device
	USBDevices     []usb.Device
	FastRPCDevices []fastrpc.Device
}

func Get(h host.Host, friendlyNames bool) (*Machine, []string, error) {
	var machineInfo Machine

	memoryInfo, err := memory.Info(h)
	if err != nil {
		return nil, nil, fmt.Errorf("getting memory info: %w", err)
	}
	machineInfo.Memory = memoryInfo

	cpus, err := cpu.Info(h)
	if err != nil {
		return nil, nil, fmt.Errorf("getting cpu info: %w", err)
	}
	machineInfo.CPUs = cpus

	diskInfo, err := disk.Info(h)
	if err != nil {
		return nil, nil, fmt.Errorf("getting disk info: %w", err)
	}
	machineInfo.Disk = diskInfo

	devices, warnings, err := device.Devices(h, friendlyNames)
	if err != nil {
		return nil, nil, fmt.Errorf("getting devices: %w", err)
	}

	// Separate devices by type
	for _, dev := range devices {
		switch d := dev.(type) {
		case pci.Device:
			machineInfo.PCIDevices = append(machineInfo.PCIDevices, d)
		case usb.Device:
			machineInfo.USBDevices = append(machineInfo.USBDevices, d)
		case fastrpc.Device:
			machineInfo.FastRPCDevices = append(machineInfo.FastRPCDevices, d)
		default:
			return nil, nil, fmt.Errorf("unknown device type: %T", dev)
		}
	}

	return &machineInfo, warnings, nil
}
