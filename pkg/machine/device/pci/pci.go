package pci

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/host"
)

const (
	BusName = "pci"

	vendorAmd    = uint16(0x1002)
	vendorIntel  = uint16(0x8086)
	vendorNvidia = uint16(0x10de)
)

// Device represents a single PCI device detected on the system.
type Device struct {
	Bus string

	Slot                 string
	BusNumber            uint8
	DeviceClass          uint32
	ProgrammingInterface *uint8
	VendorId             uint16
	DeviceId             uint16
	SubvendorId          *uint16
	SubdeviceId          *uint16
	FriendlyNames

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string
}

// FriendlyNames holds human-readable names resolved from the pci.ids database.
type FriendlyNames struct {
	VendorName    string
	DeviceName    string
	SubvendorName string
	SubdeviceName string
}

// IsGpu reports whether the device is a GPU or display controller by PCI class.
// Covers legacy VGA (0x0001) and the full display-controller class (0x03xx).
func (d Device) IsGpu() bool {
	return d.DeviceClass == 0x0001 || d.DeviceClass&0xFF00 == 0x0300
}

// pci is the PCI bus implementation.
type pci struct {
	host host.Host
	opts Options
}

// Options holds PCI-specific bus configuration.
type Options struct {
	FriendlyNames bool
}

// NewBus returns a pci bus configured with the given options.
func NewBus(targetHost host.Host, opts Options) *pci {
	return &pci{host: targetHost, opts: opts}
}

// Devices discovers all devices on the bus and returns them as a slice of Device, along with any warnings and a hard error if the bus could not be enumerated.
func (bus *pci) Devices() ([]Device, []string, error) {
	devices, warnings, err := readSysPci(bus.host)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sysfs pci devices: %w", err)
	}

	if bus.opts.FriendlyNames {
		for i, device := range devices {
			names, err := lookupFriendlyNames(bus.host, device)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("unable to get friendly name for pci device: %s", err))
			} else {
				devices[i].FriendlyNames = names
			}
		}
	}

	devices, additionalPropWarnings := addAdditionalProperties(bus.host, devices)
	warnings = append(warnings, additionalPropWarnings...)

	for i := range devices {
		devices[i].Bus = BusName
	}
	return devices, warnings, nil
}
