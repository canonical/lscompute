package usb

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "usb"

// Device represents a single USB device detected on the system.
type Device struct {
	Bus string

	BusNumber    uint8
	DeviceNumber uint8
	VendorId     uint16
	ProductId    uint16
	FriendlyNames

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string
}

// FriendlyNames holds human-readable names resolved from the usb.ids database.
type FriendlyNames struct {
	VendorName  string
	ProductName string
}

// usb is the USB bus implementation.
type usb struct {
	host host.Host
	opts Options
}

// Options holds USB-specific scanner configuration.
type Options struct {
	FriendlyNames bool
}

// NewBus returns a USB bus configured with the given options.
func NewBus(host host.Host, opts Options) *usb {
	return &usb{host: host, opts: opts}
}

// Devices discovers all USB devices on the host and returns them as Device values.
func (bus *usb) Devices() ([]Device, []string, error) {
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

	for i := range devices {
		devices[i].Bus = BusName
	}

	return devices, warnings, nil
}
