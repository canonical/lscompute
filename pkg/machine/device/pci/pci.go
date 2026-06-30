package pci

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

const (
	BusName = "pci"

	vendorAmd    = 0x1002
	vendorIntel  = 0x8086
	vendorNvidia = 0x10de
)

// Device represents a single PCI device detected on the system.
type Device struct {
	Bus string `json:"bus" yaml:"bus"`

	Slot                 string        `json:"slot" yaml:"slot"`
	BusNumber            types.HexInt  `json:"bus-number" yaml:"bus-number"`
	DeviceClass          types.HexInt  `json:"device-class" yaml:"device-class"`
	ProgrammingInterface *uint8        `json:"programming-interface,omitempty" yaml:"programming-interface,omitempty"`
	VendorId             types.HexInt  `json:"vendor-id" yaml:"vendor-id"`
	DeviceId             types.HexInt  `json:"device-id" yaml:"device-id"`
	SubvendorId          *types.HexInt `json:"subvendor-id,omitempty" yaml:"subvendor-id,omitempty"`
	SubdeviceId          *types.HexInt `json:"subdevice-id,omitempty" yaml:"subdevice-id,omitempty"`
	FriendlyNames        `json:",inline" yaml:",inline"`

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty" yaml:"additional-properties,omitempty"`
}

// FriendlyNames holds human-readable names resolved from the pci.ids database.
type FriendlyNames struct {
	VendorName    *string `json:"vendor-name,omitempty" yaml:"vendor-name,omitempty"`
	DeviceName    *string `json:"device-name,omitempty" yaml:"device-name,omitempty"`
	SubvendorName *string `json:"subvendor-name,omitempty" yaml:"subvendor-name,omitempty"`
	SubdeviceName *string `json:"subdevice-name,omitempty" yaml:"subdevice-name,omitempty"`
}

// IsGpu reports whether the device is a GPU or display controller by PCI class.
// Covers legacy VGA (0x0001) and the full display-controller class (0x03xx).
func (d Device) IsGpu() bool {
	return d.DeviceClass == 0x0001 || d.DeviceClass&0xFF00 == 0x0300
}

// pci implements bus.Bus for the PCI bus.
type pci struct {
	host host.Host
	opts Options
}

// Options holds PCI-specific bus configuration.
type Options struct {
	FriendlyNames bool
}

// NewBus returns a pci bus configured with the given options.
func NewBus(targetHost host.Host, opts Options) bus.Bus {
	return &pci{host: targetHost, opts: opts}
}

// Devices discovers all devices on the bus and returns them as a slice of any, along with any warnings and a hard error if the bus could not be enumerated.
func (bus *pci) Devices() ([]any, []string, error) {
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

	var result []any
	for _, device := range devices {
		device.Bus = BusName
		result = append(result, device)
	}
	return result, warnings, nil
}

func Decode(bytes []byte) (*Device, error) {
	var device Device
	if err := json.Unmarshal(bytes, &device); err != nil {
		return nil, err
	}
	return &device, nil
}
