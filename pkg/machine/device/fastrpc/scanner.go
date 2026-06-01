package fastrpc

import (
	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "fastrpc"

// Options holds FastRPC-specific scanner configuration.
type Options struct {
	// Future: Domains []int, etc.
}

// Scanner implements bus.Scanner for the FastRPC bus.
type Scanner struct {
	opts Options
}

// NewScanner returns a FastRPC Scanner configured with the given options.
func NewScanner(opts Options) *Scanner {
	return &Scanner{opts: opts}
}

// BusName returns the canonical FastRPC bus name.
func (s *Scanner) BusName() string { return BusName }

// Scan discovers all FastRPC devices on the host. Not yet implemented.
func (s *Scanner) Scan(h host.Host) ([]bus.DeviceInfo, []string, error) {
	// TODO: implement FastRPC device enumeration via /sys/bus/platform or /dev/fastrpc*
	return nil, nil, nil
}
