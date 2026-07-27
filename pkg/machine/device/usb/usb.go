package usb

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "usb"

// Device represents a single USB device detected on the system.
type USBDevice struct {
	Bus string `json:"bus" yaml:"bus"`

	BusNumber    int
	DeviceNumber int
	VendorId     uint64
	ProductId    uint64
	FriendlyNames

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string
}

// FriendlyNames holds human-readable names resolved from the usb.ids database.
type FriendlyNames struct {
	VendorName  string
	ProductName string
}

// usb implements bus.Bus for the USB bus.
type usb struct {
	host host.Host
	opts Options
}

// Options holds USB-specific scanner configuration.
type Options struct {
	FriendlyNames bool
}

// NewBus returns a USB bus configured with the given options.
func NewBus(host host.Host, opts Options) bus.Bus {
	return &usb{host: host, opts: opts}
}

// Devices discovers all USB devices on the host and returns them as DeviceInfo values.
func (bus *usb) Devices() ([]any, []string, error) {
	devices, warnings, err := readSysUsb(bus.host)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sysfs usb devices: %w", err)
	}

	if bus.opts.FriendlyNames {
		for i, device := range devices {
			updated, err := lookupFriendlyNames(bus.host, device)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("usb ids lookup for %04x:%04x: %v", uint64(device.VendorId), uint64(device.ProductId), err))
				continue
			}
			devices[i] = updated
		}
	}

	var result []any
	for _, device := range devices {
		device.Bus = BusName
		result = append(result, device)
	}
	return result, warnings, nil
}
