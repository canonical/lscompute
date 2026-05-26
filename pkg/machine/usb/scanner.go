package usb

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// Options holds USB-specific scanner configuration.
type Options struct {
	FriendlyNames bool
	// Future: VendorIDs []uint16, etc.
}

// Scanner implements bus.Scanner for the USB bus.
type Scanner struct {
	opts Options
}

// NewScanner returns a USB Scanner configured with the given options.
func NewScanner(opts Options) *Scanner {
	return &Scanner{opts: opts}
}

// BusName returns the canonical USB bus name.
func (s *Scanner) BusName() string { return constants.BusUsb }

// Scan discovers all USB devices on the host and returns them as DeviceInfo values.
func (s *Scanner) Scan(h host.Host) ([]types.DeviceInfo, []string, error) {
	devices, warnings, err := readSysUsb(h)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sysfs usb devices: %v", err)
	}

	if s.opts.FriendlyNames {
		for i, device := range devices {
			updated, err := lookupFriendlyNames(h, device)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("usb ids lookup for %04x:%04x: %v", uint64(device.VendorId), uint64(device.ProductId), err))
				continue
			}
			devices[i] = updated
		}
	}

	result := make([]types.DeviceInfo, len(devices))
	for i := range devices {
		d := devices[i]
		result[i] = types.DeviceInfo{Bus: constants.BusUsb, Payload: &d}
	}
	return result, warnings, nil
}
