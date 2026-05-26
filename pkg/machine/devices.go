package machine

import (
	"encoding/json"

	"github.com/canonical/lscompute/pkg/machine/bus"
	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/pci"
	"github.com/canonical/lscompute/pkg/machine/types"
	"github.com/canonical/lscompute/pkg/machine/usb"
)

func init() {
	types.RegisterBusDecoder(constants.BusPci, func(data []byte) (types.BusDevice, error) {
		var dev pci.Device
		return &dev, json.Unmarshal(data, &dev)
	})
	types.RegisterBusDecoder(constants.BusUsb, func(data []byte) (types.BusDevice, error) {
		var dev usb.Device
		return &dev, json.Unmarshal(data, &dev)
	})
	types.RegisterBusDecoder(constants.BusFastRpc, func(data []byte) (types.BusDevice, error) {
		var dev fastrpc.Device
		return &dev, json.Unmarshal(data, &dev)
	})
}

// Devices iterates all registered bus scanners and returns the combined device list.
// To add a new bus: add its NewScanner() to the scanners slice below and add a
// RegisterBusDecoder call in init() above.
func Devices(h host.Host, friendlyNames bool) ([]types.DeviceInfo, []string, error) {
	scanners := []bus.Scanner{
		pci.NewScanner(pci.Options{FriendlyNames: friendlyNames}),
		usb.NewScanner(usb.Options{FriendlyNames: friendlyNames}),
		fastrpc.NewScanner(fastrpc.Options{}),
	}

	var devices []types.DeviceInfo
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
