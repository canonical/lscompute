package machine

import (
	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/pci"
	"github.com/canonical/lscompute/pkg/machine/types"
	"github.com/canonical/lscompute/pkg/machine/usb"
)

func Devices(h host.Host, friendlyNames bool) ([]types.DeviceInfo, []string, error) {
	var machineDevices []types.DeviceInfo
	var warnings []string

	// PCI bus
	pciDevices, pciWarnings, err := pci.Devices(h, friendlyNames)
	if err != nil {
		return nil, warnings, err
	}
	for _, device := range pciDevices {
		machineDevices = append(machineDevices, types.DeviceInfo{Bus: constants.BusPci, PciDevice: device})
	}
	warnings = append(warnings, pciWarnings...)

	// USB bus
	usbDevices, usbWarnings, err := usb.Devices(h, friendlyNames)
	if err != nil {
		return nil, warnings, err
	}
	for _, device := range usbDevices {
		machineDevices = append(machineDevices, types.DeviceInfo{Bus: constants.BusUsb, UsbDevice: device})
	}
	warnings = append(warnings, usbWarnings...)

	// TODO fastrpc, etc

	return machineDevices, warnings, nil
}
