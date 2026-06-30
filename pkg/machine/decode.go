package machine

import (
	"encoding/json"
	"fmt"

	"go.yaml.in/yaml/v4"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// DecodeJSON decodes machine info from a JSON document and explicitly decodes
// each device payload using device.DecodeJSON.
func DecodeJSON(data []byte) (*MachineInfo, error) {
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
		dev, err := device.DecodeJSON(raw)
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

// DecodeYAML decodes machine info from a YAML document and explicitly decodes
// each device payload using device.DecodeYAML.
func DecodeYAML(data []byte) (*MachineInfo, error) {
	var wire struct {
		Cpus    []cpu.CpuInfo           `yaml:"cpus,omitempty"`
		Memory  memory.MemoryInfo       `yaml:"memory,omitempty"`
		Disk    map[string]disk.DirInfo `yaml:"disk,omitempty"`
		Devices []yaml.Node             `yaml:"devices,omitempty"`
	}
	if err := yaml.Unmarshal(data, &wire); err != nil {
		return nil, err
	}

	decodedDevices := make([]any, 0, len(wire.Devices))
	for i := range wire.Devices {
		dev, err := device.DecodeYAML(&wire.Devices[i])
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

