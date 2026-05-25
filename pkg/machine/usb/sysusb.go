package usb

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/types"
)

const usbDevicesPath = "/sys/bus/usb/devices"

func hostSysUsb() ([]types.UsbDevice, []string, error) {
	entries, err := os.ReadDir(usbDevicesPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", usbDevicesPath, err)
	}

	var devices []types.UsbDevice
	var warnings []string

	for _, entry := range entries {
		name := entry.Name()
		// Skip interface entries such as "2-3:1.0" — identified by the colon separator
		if strings.Contains(name, ":") {
			continue
		}

		dir := filepath.Join(usbDevicesPath, name)
		device, err := readSysUsbDevice(dir)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("reading usb device %s: %v", name, err))
			continue
		}
		devices = append(devices, device)
	}

	return devices, warnings, nil
}

func readSysUsbDevice(dir string) (types.UsbDevice, error) {
	var device types.UsbDevice

	vendorStr, err := readTrimmedFile(filepath.Join(dir, "idVendor"))
	if err != nil {
		return device, fmt.Errorf("idVendor: %w", err)
	}
	vendorId, err := strconv.ParseUint(vendorStr, 16, 16)
	if err != nil {
		return device, fmt.Errorf("parsing idVendor %q: %w", vendorStr, err)
	}
	device.VendorId = types.HexInt(vendorId)

	productStr, err := readTrimmedFile(filepath.Join(dir, "idProduct"))
	if err != nil {
		return device, fmt.Errorf("idProduct: %w", err)
	}
	productId, err := strconv.ParseUint(productStr, 16, 16)
	if err != nil {
		return device, fmt.Errorf("parsing idProduct %q: %w", productStr, err)
	}
	device.ProductId = types.HexInt(productId)

	busStr, err := readTrimmedFile(filepath.Join(dir, "busnum"))
	if err != nil {
		return device, fmt.Errorf("busnum: %w", err)
	}
	busNum, err := strconv.Atoi(busStr)
	if err != nil {
		return device, fmt.Errorf("parsing busnum %q: %w", busStr, err)
	}
	device.BusNumber = busNum

	devStr, err := readTrimmedFile(filepath.Join(dir, "devnum"))
	if err != nil {
		return device, fmt.Errorf("devnum: %w", err)
	}
	devNum, err := strconv.Atoi(devStr)
	if err != nil {
		return device, fmt.Errorf("parsing devnum %q: %w", devStr, err)
	}
	device.DeviceNumber = devNum

	return device, nil
}

func readTrimmedFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
