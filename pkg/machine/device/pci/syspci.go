package pci

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
)

const pciDevicesDir = "sys/bus/pci/devices" // io/fs path (no leading slash)

func readSysPci(h host.Host) ([]PCIDevice, []string, error) {
	entries, err := fs.ReadDir(h.FS(), pciDevicesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", pciDevicesDir, err)
	}

	var devices []PCIDevice
	var warnings []string

	for _, entry := range entries {
		slot := entry.Name()
		dir := filepath.Join(pciDevicesDir, slot)

		device, err := readSysPciDevice(h, dir, slot)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("reading pci device %s: %v", slot, err))
			continue
		}
		devices = append(devices, device)
	}

	return devices, warnings, nil
}

func readSysPciDevice(h host.Host, dir, slot string) (PCIDevice, error) {
	var device PCIDevice
	device.Slot = slot

	// slot format: "0000:3b:00.0" — index 1 is the bus number in hex
	parts := strings.Split(slot, ":")
	if len(parts) != 3 {
		return device, fmt.Errorf("unexpected slot format: %s", slot)
	}
	busNum, err := strconv.ParseUint(parts[1], 16, 8)
	if err != nil {
		return device, fmt.Errorf("parsing bus number from %q: %w", slot, err)
	}
	device.BusNumber = busNum

	vendor, err := readHexFSFile(h, filepath.Join(dir, "vendor"))
	if err != nil {
		return device, fmt.Errorf("vendor: %w", err)
	}
	device.VendorId = vendor

	deviceId, err := readHexFSFile(h, filepath.Join(dir, "device"))
	if err != nil {
		return device, fmt.Errorf("device: %w", err)
	}
	device.DeviceId = deviceId

	// class is 24-bit 0xCCSSPP: upper 16 bits are the device class (class+subclass),
	// lower 8 bits are the programming interface.
	classVal, err := readHexFSFile(h, filepath.Join(dir, "class"))
	if err != nil {
		return device, fmt.Errorf("class: %w", err)
	}
	device.DeviceClass = classVal >> 8
	if progIf := uint8(classVal & 0xFF); progIf != 0 {
		device.ProgrammingInterface = &progIf
	}

	if subVendor, err := readHexFSFile(h, filepath.Join(dir, "subsystem_vendor")); err == nil {
		device.SubvendorId = &subVendor
	}

	if subDevice, err := readHexFSFile(h, filepath.Join(dir, "subsystem_device")); err == nil {
		device.SubdeviceId = &subDevice
	}

	return device, nil
}

func readHexFSFile(h host.Host, path string) (uint64, error) {
	data, err := fs.ReadFile(h.FS(), path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	s = strings.TrimPrefix(s, "0x")
	return strconv.ParseUint(s, 16, 64)
}
