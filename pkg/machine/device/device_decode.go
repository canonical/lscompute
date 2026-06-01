package device

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

// DecodeDeviceInfo decodes a flat device JSON object by looking at its "bus" key
// and unmarshalling the payload into the corresponding bus-specific type.
func DecodeDeviceInfo(data []byte) (bus.DeviceInfo, error) {
	var peek struct {
		Bus string `json:"bus"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return bus.DeviceInfo{}, err
	}

	switch peek.Bus {
	case pci.BusName:
		var dev pci.Device
		if err := json.Unmarshal(data, &dev); err != nil {
			return bus.DeviceInfo{}, fmt.Errorf("decoding pci device: %w", err)
		}
		return bus.DeviceInfo{Bus: pci.BusName, Payload: &dev}, nil
	case usb.BusName:
		var dev usb.Device
		if err := json.Unmarshal(data, &dev); err != nil {
			return bus.DeviceInfo{}, fmt.Errorf("decoding usb device: %w", err)
		}
		return bus.DeviceInfo{Bus: usb.BusName, Payload: &dev}, nil
	case fastrpc.BusName:
		var dev fastrpc.Device
		if err := json.Unmarshal(data, &dev); err != nil {
			return bus.DeviceInfo{}, fmt.Errorf("decoding fastrpc device: %w", err)
		}
		return bus.DeviceInfo{Bus: fastrpc.BusName, Payload: &dev}, nil
	default:
		return bus.DeviceInfo{}, fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
}
