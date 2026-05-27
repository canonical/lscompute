package bus

import (
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// Scanner is the single contract a new bus package must satisfy.
// All configuration is passed at construction time via the bus-specific
// Options struct and NewScanner() — nothing is passed at scan time.
type Scanner interface {
	// BusName returns the canonical name of the bus (e.g. "pci").
	BusName() string

	// Scan returns all computation-relevant devices found on this bus.
	// Warnings are non-fatal diagnostics. A hard error means the bus
	// could not be enumerated.
	Scan(h host.Host) ([]types.DeviceInfo, []string, error)
}
