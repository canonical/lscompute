package bus

// Bus is the single contract a new bus package must satisfy.
// All configuration is passed at construction time via the bus-specific
// Options struct and NewBus() — nothing is passed at scan time.
type Bus interface {
	// Devices returns all computation-relevant devices found on this bus.
	// Warnings are non-fatal diagnostics. A hard error means the bus
	// could not be enumerated.
	Devices() ([]any, []string, error)
}
