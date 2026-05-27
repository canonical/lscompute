package types

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/constants"
)

type DeviceInfo struct {
	Bus string `json:"bus"`
	PciDevice
	UsbDevice
	FastRpc
	// Other buses
}

// busOnly is used to peek at the Bus field before full unmarshalling.
type busOnly struct {
	Bus string `json:"bus"`
}

// MarshalJSON serialises only the bus-specific sub-struct together with the Bus
// field, so that PCI zero-values are not leaked into USB output and vice-versa.
func (d DeviceInfo) MarshalJSON() ([]byte, error) {
	switch d.Bus {
	case constants.BusPci:
		return json.Marshal(struct {
			Bus string `json:"bus"`
			PciDevice
		}{d.Bus, d.PciDevice})
	case constants.BusUsb:
		return json.Marshal(struct {
			Bus string `json:"bus"`
			UsbDevice
		}{d.Bus, d.UsbDevice})
	case constants.BusFastRpc:
		return json.Marshal(struct {
			Bus string `json:"bus"`
			FastRpc
		}{d.Bus, d.FastRpc})
	default:
		return nil, fmt.Errorf("unknown device bus %q", d.Bus)
	}
}

// UnmarshalJSON selects the correct inline child struct based on the Bus field.
func (d *DeviceInfo) UnmarshalJSON(data []byte) error {
	var peek busOnly
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}
	d.Bus = peek.Bus

	switch peek.Bus {
	case constants.BusPci:
		return json.Unmarshal(data, &d.PciDevice)
	case constants.BusUsb:
		return json.Unmarshal(data, &d.UsbDevice)
	case constants.BusFastRpc:
		return json.Unmarshal(data, &d.FastRpc)
	default:
		return fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
}
