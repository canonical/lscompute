package fastrpc

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strconv"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "fastrpc"

const (
	fastRPCDevDir           = "dev"
	fastRPCDeviceNamePrefix = "fastrpc-"
)

type FastRPCDomain string

const (
	ADSPDomain FastRPCDomain = "adsp"
	MDSPDomain FastRPCDomain = "mdsp"
	SDSPDomain FastRPCDomain = "sdsp"
	CDSPDomain FastRPCDomain = "cdsp"
	GDSPDomain FastRPCDomain = "gdsp"
)

// Device represents a single FastRPC device detected on the system.
type Device struct {
	Bus string `json:"bus"`

	Domain FastRPCDomain `json:"domain"`
	Index  int           `json:"index"`
	Secure bool          `json:"secure"`

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

// Devices discovers all FastRPC devices on the host.
func (bus *fastRpc) Devices() ([]any, []string, error) {
	entries, err := fs.ReadDir(bus.host.FS(), fastRPCDevDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	result := make([]any, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, fastRPCDeviceNamePrefix) {
			continue
		}
		device, ok := parseFastRPCDeviceName(name)
		if !ok {
			continue
		}
		device.Bus = BusName
		result = append(result, device)
	}

	return result, nil, nil
}

func Decode(bytes []byte) (Device, error) {
	var device Device
	if err := json.Unmarshal(bytes, &device); err != nil {
		return Device{}, err
	}
	return device, nil
}

func parseFastRPCDeviceName(name string) (Device, bool) {
	name = strings.TrimPrefix(strings.ToLower(name), fastRPCDeviceNamePrefix)

	secure := strings.HasSuffix(name, "-secure")
	if secure {
		name = strings.TrimSuffix(name, "-secure")
	}

	i := len(name)
	for i > 0 && name[i-1] >= '0' && name[i-1] <= '9' {
		i--
	}
	domainName := name[:i]
	index := 0
	if i < len(name) {
		parsedIndex, err := strconv.Atoi(name[i:])
		if err != nil {
			return Device{}, false
		}
		index = parsedIndex
	}

	domain := FastRPCDomain(domainName)
	switch domain {
	case ADSPDomain:
	case MDSPDomain:
	case SDSPDomain:
	case CDSPDomain:
	case GDSPDomain:
	default:
		return Device{}, false
	}

	return Device{
		Domain: domain,
		Index:  index,
		Secure: secure,
	}, true
}
