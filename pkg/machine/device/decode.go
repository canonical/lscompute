package device

import (
	"encoding/json"
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

// Decode decodes a flat device JSON object by looking at its "bus" key
// and unmarshalling the payload into the corresponding bus-specific type.
func Decode(data []byte) (any, error) {
	var peek struct {
		Bus string `json:"bus"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, err
	}

	switch peek.Bus {
	case pci.BusName:
		return pci.Decode(data)
	case usb.BusName:
		return usb.Decode(data)
	case fastrpc.BusName:
		return fastrpc.Decode(data)
	default:
		return nil, fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
}
