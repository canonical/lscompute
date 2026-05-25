package usb

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/types"
)

/*
Devices returns a slice of UsbDevices that are detected on the current system.
It uses the `lsusb` command to enumerate connected USB devices.
*/
func Devices(friendlyNames bool) ([]types.UsbDevice, []string, error) {
	lsUsbOutput, err := hostLsUsb()
	if err != nil {
		return nil, nil, fmt.Errorf("executing lsusb: %v", err)
	}

	devices, warnings, err := ParseLsUsb(lsUsbOutput, friendlyNames)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing lsusb output: %v", err)
	}

	return devices, warnings, nil
}
