package usb

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

const usbDevicesDir = "sys/bus/usb/devices" // io/fs path (no leading slash)

func readSysUsb(h host.Host) ([]Device, []string, error) {
	entries, err := fs.ReadDir(h.FS(), usbDevicesDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("reading %s: %w", usbDevicesDir, err)
	}

	var devices []Device
	var warnings []string

	for _, entry := range entries {
		name := entry.Name()
		// Skip interface entries such as "2-3:1.0" — identified by the colon separator
		if strings.Contains(name, ":") {
			continue
		}

		dir := filepath.Join(usbDevicesDir, name)
		device, err := readSysUsbDevice(h, dir)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("reading usb device %s: %v", name, err))
			continue
		}
		devices = append(devices, device)
	}

	return devices, warnings, nil
}

func readSysUsbDevice(h host.Host, dir string) (Device, error) {
	var device Device

	vendorStr, err := readTrimmedFSFile(h, filepath.Join(dir, "idVendor"))
	if err != nil {
		return device, fmt.Errorf("idVendor: %w", err)
	}
	vendorId, err := strconv.ParseUint(vendorStr, 16, 16)
	if err != nil {
		return device, fmt.Errorf("parsing idVendor %q: %w", vendorStr, err)
	}
	device.VendorId = types.HexInt(vendorId)

	productStr, err := readTrimmedFSFile(h, filepath.Join(dir, "idProduct"))
	if err != nil {
		return device, fmt.Errorf("idProduct: %w", err)
	}
	productId, err := strconv.ParseUint(productStr, 16, 16)
	if err != nil {
		return device, fmt.Errorf("parsing idProduct %q: %w", productStr, err)
	}
	device.ProductId = types.HexInt(productId)

	busStr, err := readTrimmedFSFile(h, filepath.Join(dir, "busnum"))
	if err != nil {
		return device, fmt.Errorf("busnum: %w", err)
	}
	busNum, err := strconv.Atoi(busStr)
	if err != nil {
		return device, fmt.Errorf("parsing busnum %q: %w", busStr, err)
	}
	device.BusNumber = busNum

	devStr, err := readTrimmedFSFile(h, filepath.Join(dir, "devnum"))
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

func readTrimmedFSFile(h host.Host, path string) (string, error) {
	data, err := fs.ReadFile(h.FS(), path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
