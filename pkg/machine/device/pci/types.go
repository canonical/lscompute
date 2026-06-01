package pci

import (
	"github.com/canonical/lscompute/pkg/machine/types"
)

// Device represents a single PCI device detected on the system.
type Device struct {
	Slot                 string        `json:"slot"`
	BusNumber            types.HexInt  `json:"bus-number"`
	DeviceClass          types.HexInt  `json:"device-class"`
	ProgrammingInterface *uint8        `json:"programming-interface,omitempty"`
	VendorId             types.HexInt  `json:"vendor-id"`
	DeviceId             types.HexInt  `json:"device-id"`
	SubvendorId          *types.HexInt `json:"subvendor-id,omitempty"`
	SubdeviceId          *types.HexInt `json:"subdevice-id,omitempty"`
	FriendlyNames        `json:",inline"`

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

// BusName satisfies the bus.BusDevice interface.
func (d *Device) BusName() string { return BusName }

// IsGpu reports whether the device is a GPU or display controller by PCI class.
// Covers legacy VGA (0x0001) and the full display-controller class (0x03xx).
func (d Device) IsGpu() bool {
	return d.DeviceClass == 0x0001 || d.DeviceClass&0xFF00 == 0x0300
}

// FriendlyNames holds human-readable names resolved from the pci.ids database.
type FriendlyNames struct {
	VendorName    *string `json:"vendor-name,omitempty"`
	DeviceName    *string `json:"device-name,omitempty"`
	SubvendorName *string `json:"subvendor-name,omitempty"`
	SubdeviceName *string `json:"subdevice-name,omitempty"`
}
