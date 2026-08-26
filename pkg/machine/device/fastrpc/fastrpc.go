package fastrpc

import (
	"errors"
	"io/fs"
	"strconv"
	"strings"

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
	Bus string

	Domain FastRPCDomain
	Index  int
	Secure bool

	// Vendor specific device key-value pairs
	AdditionalProperties map[string]string
}

// fastRpc is the FastRPC bus implementation.
type fastRpc struct {
	host host.Host
	opts Options
}

// Options holds FastRPC-specific scanner configuration.
type Options struct {
	// e.g. FriendlyName
}

// NewBus returns a FastRPC bus configured with the given options.
func NewBus(host host.Host, opts Options) *fastRpc {
	return &fastRpc{host: host, opts: opts}
}

// Devices discovers all FastRPC devices on the host.
func (bus *fastRpc) Devices() ([]Device, []string, error) {
	entries, err := fs.ReadDir(bus.host.FS(), fastRPCDevDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	result := make([]Device, 0, len(entries))
	for _, entry := range entries {
		device, ok := parseFastRPCDeviceName(entry.Name())
		if !ok {
			continue
		}
		device.Bus = BusName
		result = append(result, device)
	}

	return result, nil, nil
}

func parseFastRPCDeviceName(name string) (Device, bool) {
	name = strings.ToLower(name)
	if !strings.HasPrefix(name, fastRPCDeviceNamePrefix) {
		return Device{}, false
	}
	name = strings.TrimPrefix(name, fastRPCDeviceNamePrefix)

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
