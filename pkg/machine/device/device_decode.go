package device

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// DecodeDeviceInfo decodes a flat device JSON object by looking at its "bus" key
// and unmarshalling the payload into the corresponding bus-specific type.
func DecodeDeviceInfo(data []byte) (types.DeviceInfo, error) {
	var peek struct {
		Bus string `json:"bus"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return types.DeviceInfo{}, err
	}

	switch peek.Bus {
	case constants.BusPci:
		var dev pci.Device
		if err := json.Unmarshal(data, &dev); err != nil {
			return types.DeviceInfo{}, fmt.Errorf("decoding pci device: %w", err)
		}
		return types.DeviceInfo{Bus: constants.BusPci, Payload: &dev}, nil
	case constants.BusUsb:
		var dev usb.Device
		if err := json.Unmarshal(data, &dev); err != nil {
			return types.DeviceInfo{}, fmt.Errorf("decoding usb device: %w", err)
		}
		return types.DeviceInfo{Bus: constants.BusUsb, Payload: &dev}, nil
	case constants.BusFastRpc:
		var dev fastrpc.Device
		if err := json.Unmarshal(data, &dev); err != nil {
			return types.DeviceInfo{}, fmt.Errorf("decoding fastrpc device: %w", err)
		}
		return types.DeviceInfo{Bus: constants.BusFastRpc, Payload: &dev}, nil
	default:
		return types.DeviceInfo{}, fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
}
