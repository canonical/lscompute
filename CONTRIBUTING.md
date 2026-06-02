# Contributing to lscompute

## How to add a new hardware bus

Adding support for a new bus (e.g. I2C, MIPI CSI, AMBA) follows a fixed recipe.
You only touch files inside your new package directory plus **one bus registration
in `device/devices.go`** and **one decoder branch in `device/decode.go`**.

### Step 1 — Create the bus package directory

```
pkg/machine/device/<busname>/
    <busname>.go    ← BusName constant, Device struct, Options, NewBus(), Devices(), Decode()
    <anything>      ← sysfs reader, ID lookup, vendor logic, tests, …
```

The simplest bus implementation lives entirely in a single `<busname>.go` file.
Add extra files (e.g. `sys<busname>.go`, `vendor.go`) when the implementation grows.

### Step 2 — Implement `<busname>.go`

```go
package <busname>

import (
    "encoding/json"
    "fmt"

    "github.com/canonical/lscompute/pkg/machine/host"
)

const BusName = "<busname>"

// Device represents a single <BusName> device detected on the system.
type Device struct {
    Bus string `json:"bus"`

    // TODO: add bus-specific fields here

    // Optional: vendor-specific key-value pairs
    AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

// Options holds <BusName>-specific bus configuration.
type Options struct{}

// <BusName> implements bus.Bus for the <BusName> bus.
type <BusName> struct {
    host host.Host
    opts Options
}

// NewBus returns a <BusName> bus configured with the given options.
func NewBus(h host.Host, opts Options) *<BusName> {
    return &<BusName>{host: h, opts: opts}
}

// Devices discovers all devices on the bus and returns them as a slice of any,
// along with non-fatal warnings and a hard error if the bus could not be enumerated.
func (s *<BusName>) Devices() ([]any, []string, error) {
    // TODO: enumerate devices, e.g. via sysfs or ioctl
    // For each discovered device, set the Bus field before appending:
    //   device.Bus = BusName
    return nil, nil, nil
}

// Decode unmarshals a raw JSON object into a *Device.
func Decode(data []byte) (*Device, error) {
    var device Device
    if err := json.Unmarshal(data, &device); err != nil {
        return nil, fmt.Errorf("decoding <busname> device: %w", err)
    }
    return &device, nil
}
```

### Step 3 — Register the bus in `device/devices.go`

Add your bus to the `buses` slice in `pkg/machine/device/devices.go`:

```go
<busname>.NewBus(h, <busname>.Options{}),
```

### Step 4 — Register the decoder in `device/decode.go`

Add one `case` to the `switch` in `pkg/machine/device/decode.go`:

```go
case <busname>.BusName:
    return <busname>.Decode(data)
```

Also import `"github.com/canonical/lscompute/pkg/machine/device/<busname>"` in both files.
