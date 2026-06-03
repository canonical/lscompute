package device

import (
	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/host"
)

// Devices iterates all registered buses and returns the combined device list.
// To add a new bus: add its NewBus() to the buses slice below and update Decode in decode.go.
func Devices(h host.Host, friendlyNames bool) ([]any, []string, error) {
	buses := []bus.Bus{
		pci.NewBus(h, pci.Options{FriendlyNames: friendlyNames}),
		usb.NewBus(h, usb.Options{FriendlyNames: friendlyNames}),
		fastrpc.NewBus(h, fastrpc.Options{}),
	}

	var devices []any
	var warnings []string
	for _, currentBus := range buses {
		devs, warns, err := currentBus.Devices()
		if err != nil {
			return nil, warnings, err
		}
		devices = append(devices, devs...)
		warnings = append(warnings, warns...)
	}
	return devices, warnings, nil
}
