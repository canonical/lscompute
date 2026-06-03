package fastrpc

import (
	"encoding/json"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "fastrpc"

// Device represents a single FastRPC device detected on the system.
type Device struct {
	// TODO: add FastRPC-specific fields (e.g. domain, instance, subsystem)

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

// fastRpc implements bus.Bus for the FastRPC bus.
type fastRpc struct {
	host host.Host
	opts Options
}

// Options holds FastRPC-specific scanner configuration.
type Options struct {
	// e.g. FriendlyName
}

// NewBus returns a FastRPC bus configured with the given options.
func NewBus(host host.Host, opts Options) bus.Bus {
	return &fastRpc{host: host, opts: opts}
}

// Devices discovers all FastRPC devices on the host. Not yet implemented.
func (bus *fastRpc) Devices() ([]any, []string, error) {
	// TODO: implement FastRPC device enumeration via /sys/bus/platform or /dev/fastrpc*

	// TODO: Copy result into an array of type any, and set the bus field to BusName

	return nil, nil, nil
}

func Decode(bytes []byte) (*Device, error) {
	var device Device
	if err := json.Unmarshal(bytes, &device); err != nil {
		return nil, err
	}
	return &device, nil
}
