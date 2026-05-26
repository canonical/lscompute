package pci

import (
	"errors"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/pci/amd"
	"github.com/canonical/lscompute/pkg/machine/pci/intel"
	"github.com/canonical/lscompute/pkg/machine/pci/nvidia"
)

// ErrorVendorNotSupported is returned by additionalProperties when no
// vendor-specific handler exists for the device's vendor ID.
var ErrorVendorNotSupported = errors.New("vendor not supported")

// addAdditionalProperties populates vendor-specific AdditionalProperties for
// each device. Unsupported vendors are silently skipped; other errors are
// collected as warnings.
func addAdditionalProperties(h host.Host, devices []Device) ([]Device, []string) {
	var warnings []string

	for i, device := range devices {
		properties, err := additionalProperties(h, device)
		if err != nil {
			if !errors.Is(err, ErrorVendorNotSupported) {
				warnings = append(warnings, fmt.Sprintf("unable to get additional properties for pci device: %s", err))
			}
		}
		devices[i].AdditionalProperties = properties
	}

	return devices, warnings
}

// additionalProperties dispatches to the correct vendor package based on the
// device's vendor ID. Add a new case here when a new vendor is supported.
func additionalProperties(h host.Host, device Device) (map[string]string, error) {
	switch device.VendorId {
	case constants.PciVendorAmd:
		props, err := amd.AdditionalProperties(h, device.Slot, device.IsGpu())
		if err != nil {
			return nil, fmt.Errorf("AMD: %v", err)
		}
		return props, nil
	case constants.PciVendorNvidia:
		props, err := nvidia.AdditionalProperties(h, device.Slot, device.IsGpu())
		if err != nil {
			return nil, fmt.Errorf("NVIDIA: %v", err)
		}
		return props, nil
	case constants.PciVendorIntel:
		props, err := intel.AdditionalProperties(h, device.Slot, device.IsGpu())
		if err != nil {
			return nil, fmt.Errorf("Intel: %v", err)
		}
		return props, nil
	default:
		return nil, ErrorVendorNotSupported
	}
}
