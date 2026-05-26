package usb

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

/*
Devices returns a slice of UsbDevices that are detected on the current system via sysfs.
*/
func Devices(h host.Host, includeFriendlyNames bool) ([]types.UsbDevice, []string, error) {
	devices, warnings, err := readSysUsb(h)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sysfs usb devices: %v", err)
	}

	if includeFriendlyNames {
		for i, device := range devices {
			entry, err := lookupUsbIds(h, device.VendorId, device.ProductId)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("usb ids lookup for %04x:%04x: %v", uint64(device.VendorId), uint64(device.ProductId), err))
				continue
			}
			if entry.VendorName != "" {
				name := entry.VendorName
				devices[i].VendorName = &name
			}
			if entry.ProductName != "" {
				name := entry.ProductName
				devices[i].ProductName = &name
			}
		}
	}

	return devices, warnings, nil
}
