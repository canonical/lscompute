package machine

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/cpu"
	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/disk"
	"github.com/canonical/lscompute/pkg/machine/memory"
)

// Decode decodes machine info JSON and explicitly decodes each device
// payload using device.Decode.
func Decode(data []byte) (*Machine, error) {
	var wire struct {
		Cpus    []cpu.CPU
		Memory  memory.Memory
		Disk    []disk.Disk
		Devices []json.RawMessage
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, err
	}
	PCIDevices := []pci.PCIDevice{}
	USBDevices := []usb.USBDevice{}
	FastRPCDevices := []fastrpc.FastRPCDevice{}
	for _, raw := range wire.Devices {
		// Peek at the bus type to determine how to decode
		var peek struct {
			Bus string `json:"bus"`
		}
		if err := json.Unmarshal(raw, &peek); err != nil {
			return nil, fmt.Errorf("peeking at device bus: %w", err)
		}

		switch peek.Bus {
		case "pci":
			var dev pci.PCIDevice
			if err := json.Unmarshal(raw, &dev); err != nil {
				return nil, fmt.Errorf("decoding pci device: %w", err)
			}
			PCIDevices = append(PCIDevices, dev)
		case "usb":
			var dev usb.USBDevice
			if err := json.Unmarshal(raw, &dev); err != nil {
				return nil, fmt.Errorf("decoding usb device: %w", err)
			}
			USBDevices = append(USBDevices, dev)
		case "fastrpc":
			var dev fastrpc.FastRPCDevice
			if err := json.Unmarshal(raw, &dev); err != nil {
				return nil, fmt.Errorf("decoding fastrpc device: %w", err)
			}
			FastRPCDevices = append(FastRPCDevices, dev)
		default:
			return nil, fmt.Errorf("unknown device bus: %q", peek.Bus)
		}
	}

	info := Machine{
		CPUs:           wire.Cpus,
		Memory:         wire.Memory,
		Disk:           wire.Disk,
		PCIDevices:     PCIDevices,
		USBDevices:     USBDevices,
		FastRPCDevices: FastRPCDevices,
	}
	return &info, nil
}
