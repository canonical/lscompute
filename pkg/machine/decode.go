package machine

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/device"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// DecodeMachineInfo decodes machine info JSON and explicitly decodes each device
// payload using device.DecodeDeviceInfo.
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
		dev, err := device.DecodeDeviceInfo(raw)
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
