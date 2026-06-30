package machine

import (
	"encoding/json"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

func TestDecodeMachineInfo(t *testing.T) {
	wire := map[string]any{
		"cpus":   []map[string]any{{"architecture": "amd64", "manufacturer-id": "GenuineIntel", "flags": []string{"sse2"}}},
		"memory": map[string]any{"total-ram": 1024, "total-swap": 0},
		"disk":   map[string]any{"/var/lib/snapd/snaps": map[string]any{"total": 100, "avail": 50}},
		"devices": []map[string]any{
			{"bus": "usb", "bus-number": 1, "device-number": 2, "vendor-id": "0x0bda", "product-id": "0x5487"},
			{"bus": "pci", "slot": "0000:00:02.0", "bus-number": "0x00", "device-class": "0x0300", "vendor-id": "0x8086", "device-id": "0x5916"},
		},
	}

	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	info, err := DecodeJSON(data)
	if err != nil {
		t.Fatalf("DecodeMachineInfo() error: %v", err)
	}
	if len(info.Devices) != 2 {
		t.Fatalf("len(Devices) = %d, want 2", len(info.Devices))
	}
	if _, ok := info.Devices[0].(usb.Device); !ok {
		t.Fatalf("Devices[0] type = %T, want usb.Device", info.Devices[0])
	}
	if _, ok := info.Devices[1].(pci.Device); !ok {
		t.Fatalf("Devices[1] type = %T, want pci.Device", info.Devices[1])
	}
}

func TestDecodeMachineInfo_InvalidDevice(t *testing.T) {
	// Build JSON directly with an unknown bus — DecodeMachineInfo must return an error.
	data := []byte(`{"devices":[{"bus":"unknown","vendor-id":1}]}`)
	if _, err := DecodeJSON(data); err == nil {
		t.Fatal("expected error for unknown bus, got nil")
	}
}

func TestDecodeMachineInfo_MalformedJSON(t *testing.T) {
	_, err := DecodeJSON([]byte(`not valid json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestDecodeMachineInfo_YAML(t *testing.T) {
	data := []byte(`
cpus:
  - architecture: amd64
    manufacturer-id: GenuineIntel
    flags:
      - sse2
memory:
  total-ram: 1024
  total-swap: 0
disk:
  /var/lib/snapd/snaps:
    total: 100
    avail: 50
devices:
  - bus: usb
    bus-number: 1
    device-number: 2
    vendor-id: "0x0bda"
    product-id: "0x5487"
  - bus: pci
    slot: "0000:00:02.0"
    bus-number: "0x00"
    device-class: "0x0300"
    vendor-id: "0x8086"
    device-id: "0x5916"
`)

	info, err := DecodeYAML(data)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if len(info.Cpus) != 1 || info.Cpus[0].Architecture != "amd64" {
		t.Fatalf("unexpected cpus: %+v", info.Cpus)
	}
	if info.Memory.TotalRam != 1024 {
		t.Fatalf("Memory.TotalRam = %d, want 1024", info.Memory.TotalRam)
	}
	if len(info.Devices) != 2 {
		t.Fatalf("len(Devices) = %d, want 2", len(info.Devices))
	}
	if _, ok := info.Devices[0].(usb.Device); !ok {
		t.Fatalf("Devices[0] type = %T, want usb.Device", info.Devices[0])
	}
	pciDev, ok := info.Devices[1].(pci.Device)
	if !ok {
		t.Fatalf("Devices[1] type = %T, want pci.Device", info.Devices[1])
	}
	if int(pciDev.VendorId) != 0x8086 {
		t.Fatalf("pci VendorId = 0x%x, want 0x8086", int(pciDev.VendorId))
	}
}

func TestDecodeMachineInfo_YAMLUnknownBus(t *testing.T) {
	data := []byte("devices:\n  - bus: unknown\n    vendor-id: 1\n")
	if _, err := DecodeYAML(data); err == nil {
		t.Fatal("expected error for unknown bus, got nil")
	}
}

