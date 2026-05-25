package machine

import (
	"github.com/canonical/lscompute/pkg/machine/pci"
	"github.com/canonical/lscompute/pkg/machine/types"
	"github.com/canonical/lscompute/pkg/machine/usb"
)

func Devices(friendlyNames bool) ([]types.DeviceInfo, []string, error) {
	var machineDevices []types.DeviceInfo
	var warnings []string

	// PCI bus
	pciDevices, pciWarnings, err := pci.Devices(friendlyNames)
	if err != nil {
		return nil, warnings, err
	}
	for _, device := range pciDevices {
		machineDevices = append(machineDevices, types.DeviceInfo{Bus: "pci", PciDevice: device})
	}
	warnings = append(warnings, pciWarnings...)

	// USB bus
	usbDevices, usbWarnings, err := usb.Devices(friendlyNames)
	if err != nil {
		return nil, warnings, err
	}
	for _, device := range usbDevices {
		machineDevices = append(machineDevices, types.DeviceInfo{Bus: "usb", UsbDevice: device})
	}
	warnings = append(warnings, usbWarnings...)

	// TODO fastrpc, etc

	return machineDevices, warnings, nil
}
