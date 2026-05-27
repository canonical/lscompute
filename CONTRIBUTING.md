# Contributing to lscompute

## How to add a new hardware bus

Adding support for a new bus (e.g. I2C, MIPI CSI, AMBA) follows a fixed six-step
recipe. You only touch files inside your new package directory plus **one scanner
registration in `device/devices.go`** and **one device-decoder branch in `device/device_decode.go`**.

### Step 1 — Create the bus package directory

```
pkg/machine/device/<busname>/
    types.go     ← device struct + BusName()
    scanner.go   ← scanner + options
    <anything>   ← sysfs reader, ID lookup, vendor logic, tests, …
```

### Step 2 — Define `types.go`

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
    "github.com/canonical/lscompute/pkg/machine/device/bus"
    "github.com/canonical/lscompute/pkg/machine/constants"
    "github.com/canonical/lscompute/pkg/machine/host"
    "github.com/canonical/lscompute/pkg/machine/types"
)

// Options holds <BusName>-specific scanner configuration.
type Options struct {
    // Add bus-specific tuning fields here. These are set at construction time
    // via NewScanner().
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

### Step 4 — Register scanner in `device/devices.go`

Open `pkg/machine/device/devices.go` and add the new scanner:

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
```

### Step 5 — Register device decoder in `device/device_decode.go`

Open `pkg/machine/device/device_decode.go` and add one switch branch in
`DecodeDeviceInfo`:

```go
case constants.Bus<BusName>:
    var dev <busname>.Device
    if err := json.Unmarshal(data, &dev); err != nil {
        return types.DeviceInfo{}, fmt.Errorf("decoding <busname> device: %w", err)
    }
    return types.DeviceInfo{Bus: constants.Bus<BusName>, Payload: &dev}, nil
```

### Step 6 — Add test fixtures (recommended)

Place pre-captured sysfs or command output under:

```
test_data/machines/<machine-name>/machine-root/sys/bus/<busname>/
```

The golden-file pipeline test (`TestGetFromMachineDirs`) will pick it up automatically.

---

## Architecture overview

```
pkg/machine/
    machine.go              ← top-level Get()
    decode.go               ← DecodeMachineInfo() — JSON round-trip for MachineInfo
    device/
        bus/scanner.go      ← Scanner interface (shared contract)
        devices.go          ← scanner registry
        device_decode.go    ← DecodeDeviceInfo() — explicit JSON decode for bus payloads
        pci/
            types.go        ← pci.Device implements BusDevice + BusName()
            scanner.go      ← pci.Scanner implements bus.Scanner
            syspci.go       ← internal enumeration logic
            amd/, intel/, nvidia/  ← vendor-specific additional properties
        usb/
            types.go        ← usb.Device implements BusDevice + BusName()
            scanner.go      ← usb.Scanner implements bus.Scanner
            sysusb.go       ← internal enumeration logic
        fastrpc/
            types.go        ← fastrpc.Device implements BusDevice + BusName()
            scanner.go      ← fastrpc.Scanner (stub, not yet implemented)
    types/device.go         ← DeviceInfo + BusDevice interface
```

### Configuration model

| Scope | Where | When |
|---|---|---|
| Cross-bus (e.g. `FriendlyNames`) | `machine.Devices(..., friendlyNames)` | Passed once and forwarded to buses that support it |
| Bus-specific tuning | `<busname>.Options`, passed via `NewScanner()` | Only for one bus |
