package device

import (
	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/host"
)

// Devices iterates all registered bus scanners and returns the combined device list.
// To add a new bus: add its NewScanner() to the scanners slice below and update
// DecodeDeviceInfo in device_decode.go.
func Devices(h host.Host, friendlyNames bool) ([]bus.DeviceInfo, []string, error) {
	scanners := []bus.Scanner{
		pci.NewScanner(pci.Options{FriendlyNames: friendlyNames}),
		usb.NewScanner(usb.Options{FriendlyNames: friendlyNames}),
		fastrpc.NewScanner(fastrpc.Options{}),
	}

	var devices []bus.DeviceInfo
	var warnings []string
	for _, s := range scanners {
		devs, warns, err := s.Scan(h)
		if err != nil {
			return nil, warnings, err
		}
		devices = append(devices, devs...)
		warnings = append(warnings, warns...)
	}
	return devices, warnings, nil
}
