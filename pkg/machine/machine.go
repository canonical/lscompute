package machine

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/cpu"
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

func Get(h host.Host, friendlyNames bool, retrieveAllDevices bool) (*Machine, []string, error) {
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

	var warnings []string

	pciBus := pci.NewBus(h, pci.Options{FriendlyNames: friendlyNames})
	if d, w, err := pciBus.Devices(retrieveAllDevices); err != nil {
		return nil, nil, fmt.Errorf("getting PCI devices: %w", err)
	} else {
		machineInfo.PCIDevices = d
		warnings = append(warnings, w...)
	}

	fastRPCBus := fastrpc.NewBus(h, fastrpc.Options{})
	if d, w, err := fastRPCBus.Devices(); err != nil {
		return nil, nil, fmt.Errorf("getting FastRPC devices: %w", err)
	} else {
		machineInfo.FastRPCDevices = d
		warnings = append(warnings, w...)
	}

	if retrieveAllDevices {
		usbBus := usb.NewBus(h, usb.Options{FriendlyNames: friendlyNames})
		if d, w, err := usbBus.Devices(); err != nil {
			return nil, nil, fmt.Errorf("getting USB devices: %w", err)
		} else {
			machineInfo.USBDevices = d
			warnings = append(warnings, w...)
		}
	}

	return &machineInfo, warnings, nil
}
