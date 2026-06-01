package pci

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "pci"

// Options holds PCI-specific scanner configuration.
type Options struct {
	FriendlyNames bool
	// Future: VendorIDs []uint16, DeviceClasses []uint16, etc.
}

// Scanner implements bus.Scanner for the PCI bus.
type Scanner struct {
	opts Options
}

// NewScanner returns a PCI Scanner configured with the given options.
func NewScanner(opts Options) *Scanner {
	return &Scanner{opts: opts}
}

// Scan discovers all PCI devices on the host and returns them as DeviceInfo values.
func (s *Scanner) Scan(h host.Host) ([]bus.DeviceInfo, []string, error) {
	devices, warnings, err := readSysPci(h)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sysfs pci devices: %w", err)
	}

	if s.opts.FriendlyNames {
		for i, device := range devices {
			names, err := lookupFriendlyNames(h, device)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("unable to get friendly name for pci device: %s", err))
			} else {
				devices[i].FriendlyNames = names
			}
		}
	}

	devices, additionalPropWarnings := addAdditionalProperties(h, devices)
	warnings = append(warnings, additionalPropWarnings...)

	result := make([]bus.DeviceInfo, len(devices))
	for i := range devices {
		d := devices[i]
		result[i] = bus.DeviceInfo{Bus: BusName, Payload: &d}
	}
	return result, warnings, nil
}
