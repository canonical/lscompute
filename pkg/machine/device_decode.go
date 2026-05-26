package machine

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/pci"
	"github.com/canonical/lscompute/pkg/machine/types"
	"github.com/canonical/lscompute/pkg/machine/usb"
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

// DecodeMachineInfo decodes machine info JSON and explicitly decodes each device
// payload using DecodeDeviceInfo.
func DecodeMachineInfo(data []byte) (*types.MachineInfo, error) {
	var wire struct {
		Cpus    []types.CpuInfo           `json:"cpus,omitempty"`
		Memory  types.MemoryInfo          `json:"memory,omitempty"`
		Disk    map[string]types.DirStats `json:"disk,omitempty"`
		Devices []json.RawMessage         `json:"devices,omitempty"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, err
	}

	decodedDevices := make([]types.DeviceInfo, 0, len(wire.Devices))
	for _, raw := range wire.Devices {
		dev, err := DecodeDeviceInfo(raw)
		if err != nil {
			return nil, fmt.Errorf("decoding machine device: %w", err)
		}
		decodedDevices = append(decodedDevices, dev)
	}

	info := types.MachineInfo{
		Cpus:    wire.Cpus,
		Memory:  wire.Memory,
		Disk:    wire.Disk,
		Devices: decodedDevices,
	}
	return &info, nil
}
