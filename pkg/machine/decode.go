package machine

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// Decode decodes machine info JSON and explicitly decodes each device
// payload using device.Decode.
func Decode(data []byte) (*MachineInfo, error) {
	var wire struct {
		Cpus    []cpu.CpuInfo           `json:"cpus,omitempty"`
		Memory  memory.MemoryInfo       `json:"memory,omitempty"`
		Disk    map[string]disk.DirInfo `json:"disk,omitempty"`
		Devices []json.RawMessage       `json:"devices,omitempty"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, err
	}

	decodedDevices := make([]any, 0, len(wire.Devices))
	for _, raw := range wire.Devices {
		dev, err := device.Decode(raw)
		if err != nil {
			return nil, fmt.Errorf("decoding machine device: %w", err)
		}
		decodedDevices = append(decodedDevices, dev)
	}

	info := MachineInfo{
		Cpus:    wire.Cpus,
		Memory:  wire.Memory,
		Disk:    wire.Disk,
		Devices: decodedDevices,
	}
	return &info, nil
}
