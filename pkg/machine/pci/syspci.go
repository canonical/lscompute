package pci

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

const pciDevicesDir = "sys/bus/pci/devices" // io/fs path (no leading slash)

func readSysPci(h host.Host) ([]Device, []string, error) {
	entries, err := fs.ReadDir(h.FS(), pciDevicesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", pciDevicesDir, err)
	}

	var devices []Device
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

func readSysPciDevice(h host.Host, dir, slot string) (Device, error) {
	var device Device
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
	device.BusNumber = types.HexInt(busNum)

	vendor, err := readHexFSFile(h, filepath.Join(dir, "vendor"))
	if err != nil {
		return device, fmt.Errorf("vendor: %w", err)
	}
	device.VendorId = types.HexInt(vendor)

	deviceId, err := readHexFSFile(h, filepath.Join(dir, "device"))
	if err != nil {
		return device, fmt.Errorf("device: %w", err)
	}
	device.DeviceId = types.HexInt(deviceId)

	// class is 24-bit 0xCCSSPP: upper 16 bits are the device class (class+subclass),
	// lower 8 bits are the programming interface.
	classVal, err := readHexFSFile(h, filepath.Join(dir, "class"))
	if err != nil {
		return device, fmt.Errorf("class: %w", err)
	}
	device.DeviceClass = types.HexInt(classVal >> 8)
	if progIf := uint8(classVal & 0xFF); progIf != 0 {
		device.ProgrammingInterface = &progIf
	}

	if subVendor, err := readHexFSFile(h, filepath.Join(dir, "subsystem_vendor")); err == nil {
		sv := types.HexInt(subVendor)
		device.SubvendorId = &sv
	}

	if subDevice, err := readHexFSFile(h, filepath.Join(dir, "subsystem_device")); err == nil {
		sd := types.HexInt(subDevice)
		device.SubdeviceId = &sd
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
