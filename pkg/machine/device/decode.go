package device

import (
	"encoding/json"
	"fmt"

	"go.yaml.in/yaml/v4"

	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

// DecodeJSON decodes a flat device JSON object by looking at its "bus" key
// and unmarshalling the payload into the corresponding bus-specific type.
func DecodeJSON(data []byte) (any, error) {
	var peek struct {
		Bus string `json:"bus"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, err
	}

	switch peek.Bus {
	case pci.BusName:
		return pci.DecodeJSON(data)
	case usb.BusName:
		return usb.DecodeJSON(data)
	case fastrpc.BusName:
		return fastrpc.DecodeJSON(data)
	default:
		return nil, fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
}

// DecodeYAML decodes a device YAML node by looking at its "bus" key
// and unmarshalling the payload into the corresponding bus-specific type.
func DecodeYAML(value *yaml.Node) (any, error) {
	var peek struct {
		Bus string `yaml:"bus"`
	}
	if err := value.Decode(&peek); err != nil {
		return nil, err
	}

	switch peek.Bus {
	case pci.BusName:
		return pci.DecodeYAML(value)
	case usb.BusName:
		return usb.DecodeYAML(value)
	case fastrpc.BusName:
		return fastrpc.DecodeYAML(value)
	default:
		return nil, fmt.Errorf("unknown device bus: %q", peek.Bus)
	}
}

