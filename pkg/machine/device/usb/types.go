package usb

import (
	"github.com/canonical/lscompute/pkg/machine/types"
)

// Device represents a single USB device detected on the system.
type Device struct {
	BusNumber     int          `json:"bus-number"`
	DeviceNumber  int          `json:"device-number"`
	VendorId      types.HexInt `json:"vendor-id"`
	ProductId     types.HexInt `json:"product-id"`
	FriendlyNames `json:",inline"`

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}


// FriendlyNames holds human-readable names resolved from the usb.ids database.
type FriendlyNames struct {
	VendorName  *string `json:"vendor-name,omitempty"`
	ProductName *string `json:"product-name,omitempty"`
}
