package fastrpc

import (
	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

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
func (s *Scanner) BusName() string { return constants.BusFastRpc }

// Scan discovers all FastRPC devices on the host. Not yet implemented.
func (s *Scanner) Scan(h host.Host) ([]types.DeviceInfo, []string, error) {
	// TODO: implement FastRPC device enumeration via /sys/bus/platform or /dev/fastrpc*
	return nil, nil, nil
}
