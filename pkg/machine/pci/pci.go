package pci

import (
	"errors"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/pci/amd"
	"github.com/canonical/lscompute/pkg/machine/pci/intel"
	"github.com/canonical/lscompute/pkg/machine/pci/nvidia"
	"github.com/canonical/lscompute/pkg/machine/types"
)

var (
	ErrorVendorNotSupported = errors.New("vendor not supported")
)

/*
Devices returns a slice of PciDevices that are detected on the current system via sysfs.
*/
func Devices(includeFriendlyNames bool) ([]types.PciDevice, []string, error) {
	devices, warnings, err := hostSysPci()
	if err != nil {
		return nil, nil, fmt.Errorf("reading sysfs pci devices: %v", err)
	}

	if includeFriendlyNames {
		for i, device := range devices {
			names, err := friendlyNames(device)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("unable to get friendly name for pci device: %s", err))
			} else {
				devices[i].PciFriendlyNames = names
			}
		}
	}

	devices, additionalPropWarnings := addAdditionalProperties(devices)
	warnings = append(warnings, additionalPropWarnings...)
	return devices, warnings, nil
}


/*
friendlyNames uses the numeric PCI ID fields to look up human-readable names for the device from the pci.ids database.
*/
func friendlyNames(device types.PciDevice) (types.PciFriendlyNames, error) {
	entry, err := lookupPciIds(device.VendorId, device.DeviceId, device.SubvendorId, device.SubdeviceId)
	if err != nil {
		return types.PciFriendlyNames{}, fmt.Errorf("pci ids lookup for %04x:%04x: %w", uint64(device.VendorId), uint64(device.DeviceId), err)
	}

	var names types.PciFriendlyNames
	if entry.VendorName != "" {
		s := entry.VendorName
		names.VendorName = &s
	}
	if entry.DeviceName != "" {
		s := entry.DeviceName
		names.DeviceName = &s
	}
	if entry.SubvendorName != "" {
		s := entry.SubvendorName
		names.SubvendorName = &s
	}
	if entry.SubdeviceName != "" {
		s := entry.SubdeviceName
		names.SubdeviceName = &s
	}
	return names, nil
}

/*
addAdditionalProperties returns devices with their AdditionalProperties field populated with device specific properties.
Additional properties are obtained by running vendor specific tools on the host system.
No error is returned as a failure to look up properties is considered non-fatal, and likely due to missing drivers.
Errors are instead logged to STDERR.
*/
func addAdditionalProperties(devices []types.PciDevice) ([]types.PciDevice, []string) {
	var warnings []string

	for i, device := range devices {
		properties, err := deviceAdditionalProperties(device)
		if err != nil {
			if errors.Is(err, ErrorVendorNotSupported) {
				// We do not log unsupported vendors, as that would be the majority of PCI devices
			} else {
				warnings = append(warnings, fmt.Sprintf("unable to get additional properties for pci device: %s", err))
			}
		}
		devices[i].AdditionalProperties = properties
	}

	return devices, warnings
}

func deviceAdditionalProperties(device types.PciDevice) (map[string]string, error) {
	var properties map[string]string
	var err error

	switch device.VendorId {
	case constants.PciVendorAmd:
		properties, err = amd.AdditionalProperties(device)
		if err != nil {
			return nil, fmt.Errorf("AMD: %v", err)
		}
	case constants.PciVendorNvidia:
		properties, err = nvidia.AdditionalProperties(device)
		if err != nil {
			return nil, fmt.Errorf("NVIDIA: %v", err)
		}
	case constants.PciVendorIntel:
		properties, err = intel.AdditionalProperties(device)
		if err != nil {
			return nil, fmt.Errorf("Intel: %v", err)
		}
	default:
		return nil, ErrorVendorNotSupported
	}

	return properties, nil
}
