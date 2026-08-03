# Contributing to lscompute

## How to add a new hardware bus

Adding support for a new bus (e.g. I2C, MIPI CSI, AMBA) follows a fixed recipe.
You only add a new package inside `device` directory and load it in `machine.go`.

### Step 1 — Create the bus package directory

```
pkg/machine/device/<busname>/
    <busname>.go    ← BusName constant, Device struct, Options, NewBus(), Devices()
    <anything>      ← sysfs reader, ID lookup, vendor logic, tests, …
```

The simplest bus implementation lives entirely in a single `<busname>.go` file.
Add extra files (e.g. `sys<busname>.go`, `vendor.go`) when the implementation grows.

### Step 2 — Implement `<busname>.go`

```go
package <busname>

import (
    "github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "<busname>"

// Device represents a single <busType> device detected on the system.
//
// Public device structs carry no serialization tags: the human-readable
// JSON/YAML rendering lives in `cmd/lscompute` (see `MachineDetails`).
type Device struct {
    Bus string

    // TODO: add bus-specific fields here

    // Optional: vendor-specific key-value pairs
    AdditionalProperties map[string]string
}

// Options holds <busType>-specific bus configuration.
type Options struct{}

// <busType> is the <busType> bus implementation.
type <busType> struct {
    host host.Host
    opts Options
}

// NewBus returns a <busType> bus configured with the given options.
func NewBus(h host.Host, opts Options) *<busType> {
    return &<busType>{host: h, opts: opts}
}

// Devices discovers all <busType> devices on the host and returns them along
// with any warnings and a hard error if the bus could not be enumerated.
func (bus *<busType>) Devices() ([]Device, []string, error) {
    // TODO: enumerate devices; set device.Bus = BusName on each result
    return nil, nil, nil
}
```

### Step 3 — Add the bus in `machine.go`

Add your bus to the `Get()` method in `pkg/machine/machine.go`:

```go
<busname> := <busname>.NewBus(h, <busname>.Options{})
	if d, w, err := <busname>.Devices(); err != nil {
		return nil, nil, fmt.Errorf("getting <busname> devices: %w", err)
	} else {
		machineInfo.<BusType>Devices = d
		warnings = append(warnings, w...)
	}
```

Also import `"github.com/canonical/lscompute/pkg/machine/device/<busname>"` in `pkg/machine/machine.go`.
