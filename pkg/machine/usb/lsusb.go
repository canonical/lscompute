package usb

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/types"
)

func hostLsUsb() (string, error) {
	out, err := exec.Command("lsusb").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ParseLsUsb parses the output of `lsusb` into a slice of UsbDevice.
// Each line looks like:
//
//	Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub
func ParseLsUsb(inputString string, includeFriendlyNames bool) ([]types.UsbDevice, []string, error) {
	var devices []types.UsbDevice
	var warnings []string

	for _, line := range strings.Split(strings.TrimSpace(inputString), "\n") {
		if line == "" {
			continue
		}

		device, err := parseLsUsbLine(line)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipping usb line %q: %v", line, err))
			continue
		}

		if includeFriendlyNames {
			entry, err := lookupUsbIds(device.VendorId, device.ProductId)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("usb ids lookup for %04x:%04x: %v", uint64(device.VendorId), uint64(device.ProductId), err))
			} else {
				if entry.VendorName != "" {
					device.VendorName = &entry.VendorName
				}
				if entry.ProductName != "" {
					device.ProductName = &entry.ProductName
				}
			}
		}

		devices = append(devices, device)
	}

	return devices, warnings, nil
}

// parseLsUsbLine parses a single line of lsusb output.
// Format: Bus BBB Device DDD: ID VVVV:PPPP [description]
func parseLsUsbLine(line string) (types.UsbDevice, error) {
	var device types.UsbDevice

	// Minimum expected: "Bus 001 Device 001: ID 0000:0000"
	// Split on " ID " to get the bus/device part and the id/description part
	parts := strings.SplitN(line, " ID ", 2)
	if len(parts) != 2 {
		return device, fmt.Errorf("unexpected format: missing ' ID '")
	}

	busDevicePart := strings.TrimSpace(parts[0]) // "Bus 002 Device 001:"
	idDescPart := strings.TrimSpace(parts[1])    // "1d6b:0003 Linux Foundation 3.0 root hub"

	// Parse bus number and device number
	// busDevicePart: "Bus 002 Device 001:"
	busDevicePart = strings.TrimSuffix(busDevicePart, ":")
	fields := strings.Fields(busDevicePart)
	if len(fields) < 4 {
		return device, fmt.Errorf("unexpected format for bus/device: %q", busDevicePart)
	}
	busNum, err := strconv.Atoi(fields[1])
	if err != nil {
		return device, fmt.Errorf("cannot parse bus number %q: %v", fields[1], err)
	}
	deviceNum, err := strconv.Atoi(fields[3])
	if err != nil {
		return device, fmt.Errorf("cannot parse device number %q: %v", fields[3], err)
	}
	device.BusNumber = busNum
	device.DeviceNumber = deviceNum

	// Parse vendor:product and optional description
	// idDescPart: "1d6b:0003 Linux Foundation 3.0 root hub"
	idAndDesc := strings.SplitN(idDescPart, " ", 2)
	idPart := idAndDesc[0] // "1d6b:0003"

	idFields := strings.SplitN(idPart, ":", 2)
	if len(idFields) != 2 {
		return device, fmt.Errorf("unexpected format for vendor:product ID: %q", idPart)
	}
	vendorId, err := strconv.ParseUint(idFields[0], 16, 16)
	if err != nil {
		return device, fmt.Errorf("cannot parse vendor id %q: %v", idFields[0], err)
	}
	productId, err := strconv.ParseUint(idFields[1], 16, 16)
	if err != nil {
		return device, fmt.Errorf("cannot parse product id %q: %v", idFields[1], err)
	}
	device.VendorId = types.HexInt(vendorId)
	device.ProductId = types.HexInt(productId)

	return device, nil
}



