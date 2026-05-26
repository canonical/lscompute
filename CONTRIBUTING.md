# Contributing to lscompute

## How to add a new hardware bus

Adding support for a new bus (e.g. I2C, MIPI CSI, AMBA) follows a fixed four-step
recipe. You only touch files inside your new package directory plus **two lines in
`devices.go`**. No other existing file needs editing.

### Step 1 — Create the bus package directory

```
pkg/machine/<busname>/
    device.go    ← device struct
    scanner.go   ← scanner + options
    <anything>   ← sysfs reader, ID lookup, vendor logic, tests, …
```

### Step 2 — Define `device.go`

```go
package <busname>

import (
    "github.com/canonical/lscompute/pkg/machine/constants"
)

// Device represents a single <BusName> device detected on the system.
type Device struct {
    // TODO: add bus-specific fields

    // Optional: vendor-specific key-value pairs
    AdditionalProperties map[string]string `json:"additional-properties,omitempty"`
}

// BusName satisfies the types.BusDevice interface.
func (d *Device) BusName() string { return constants.Bus<BusName> }
```

Add `Bus<BusName> = "<busname>"` to `pkg/machine/constants/constants.go`.

### Step 3 — Define `scanner.go`

```go
package <busname>

import (
    "github.com/canonical/lscompute/pkg/machine/bus"
    "github.com/canonical/lscompute/pkg/machine/constants"
    "github.com/canonical/lscompute/pkg/machine/host"
    "github.com/canonical/lscompute/pkg/machine/types"
)

// Options holds <BusName>-specific scanner configuration.
type Options struct {
    // Add bus-specific tuning fields here. These are set at construction time
    // via NewScanner() and do NOT affect the shared bus.ScanOptions interface.
}

// Scanner implements bus.Scanner for the <BusName> bus.
type Scanner struct{ opts Options }

func NewScanner(opts Options) *Scanner { return &Scanner{opts: opts} }

func (s *Scanner) BusName() string { return constants.Bus<BusName> }

func (s *Scanner) Scan(h host.Host) ([]types.DeviceInfo, []string, error) {
    // TODO: enumerate devices, e.g. via sysfs or ioctl
    var devices []types.DeviceInfo
    return devices, nil, nil
}
```

### Step 4 — Register in `devices.go`

Open `pkg/machine/devices.go` and add **two lines**:

```go
func Devices(h host.Host, friendlyNames bool) ([]types.DeviceInfo, []string, error) {
    scanners := []bus.Scanner{
        pci.NewScanner(pci.Options{FriendlyNames: friendlyNames}),
        usb.NewScanner(usb.Options{FriendlyNames: friendlyNames}),
        fastrpc.NewScanner(fastrpc.Options{}),
        <busname>.NewScanner(<busname>.Options{}),  // ← add this
    }
    // ...
}

func init() {
    // existing registrations …
    types.RegisterBusDecoder(constants.Bus<BusName>, func(data []byte) (types.BusDevice, error) {
        var dev <busname>.Device
        return &dev, json.Unmarshal(data, &dev)
    })  // ← add this
}
```

### Step 5 — Add test fixtures (recommended)

Place pre-captured sysfs or command output under:

```
test_data/machines/<machine-name>/machine-root/sys/bus/<busname>/
```

The golden-file pipeline test (`TestGetFromMachineDirs`) will pick it up automatically.

---

## Architecture overview

```
pkg/machine/
    bus/scanner.go          ← Scanner interface + ScanOptions (shared contract)
    devices.go              ← scanner registry + decoder registrations
    machine.go              ← top-level Get()
    types/device.go         ← DeviceInfo, BusDevice interface, decoder registry
    pci/
        device.go           ← pci.Device implements BusDevice
        scanner.go          ← pci.Scanner implements bus.Scanner
        pci.go              ← internal enumeration logic
        amd/, intel/, nvidia/  ← vendor-specific additional properties
    usb/
        device.go           ← usb.Device implements BusDevice
        scanner.go          ← usb.Scanner implements bus.Scanner
        usb.go, sysusb.go   ← internal enumeration logic
    fastrpc/
        device.go           ← fastrpc.Device implements BusDevice
        scanner.go          ← fastrpc.Scanner (stub, not yet implemented)
```

### Two-level options model

| Scope | Where | When |
|---|---|---|
| Cross-bus (e.g. `FriendlyNames`) | `bus.ScanOptions`, passed via `Scan()` | Applies to all scanners |
| Bus-specific tuning | `<busname>.Options`, passed via `NewScanner()` | Only for one bus |

