# Contributing to lscompute

## How to add a new hardware bus

Adding support for a new bus (e.g. I2C, MIPI CSI, AMBA) follows a fixed recipe.
You only touch files inside your new package directory plus **one bus registration
in `device/devices.go`**.

### Step 1 — Create the bus package directory

```
pkg/machine/device/<busname>/
    <busname>.go    ← BusName constant, <BusType>Device struct, Options, NewBus(), Devices()
    <anything>      ← sysfs reader, ID lookup, vendor logic, tests, …
```

The simplest bus implementation lives entirely in a single `<busname>.go` file.
Add extra files (e.g. `sys<busname>.go`, `vendor.go`) when the implementation grows.

### Step 2 — Implement `<busname>.go`

```go
package <busname>

import (
    "github.com/canonical/lscompute/pkg/machine/device/bus"
    "github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "<busname>"

// <BusType>Device represents a single <busType> device detected on the system.
//
// Public device structs carry no serialization tags: the human-readable
// JSON/YAML rendering lives in `cmd/lscompute` (see `MachineDetails`).
type <BusType>Device struct {
    Bus string

    // TODO: add bus-specific fields here

    // Optional: vendor-specific key-value pairs
    AdditionalProperties map[string]string
}

// Options holds <busType>-specific bus configuration.
type Options struct{}

// <busType> implements bus.Bus for the <busType> bus.
type <busType> struct {
    host host.Host
    opts Options
}

// NewBus returns a <busType> bus configured with the given options.
func NewBus(h host.Host, opts Options) bus.Bus {
    return &<busType>{host: h, opts: opts}
}

// Devices discovers all devices on the bus and returns them as a slice of any,
// along with non-fatal warnings and a hard error if the bus could not be enumerated.
func (bus *<busType>) Devices() ([]any, []string, error) {
    // TODO: enumerate devices, e.g. via sysfs or ioctl
    // For each discovered device, set the Bus field before appending:
    //   device.Bus = BusName
    return nil, nil, nil
}
```

### Step 3 — Register the bus in `device/devices.go`

Add your bus to the `buses` slice in `pkg/machine/device/devices.go`:

```go
<busname>.NewBus(h, <busname>.Options{}),
```

Also import `"github.com/canonical/lscompute/pkg/machine/device/<busname>"` in `device/devices.go`.
