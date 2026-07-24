package machine

import (
	"encoding/json"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
)

func TestDecodeMachine(t *testing.T) {
	wire := map[string]any{
		"cpus":   []map[string]any{{"architecture": "amd64", "manufacturer-id": "GenuineIntel", "flags": []string{"sse2"}}},
		"memory": map[string]any{"total-ram": 1024, "total-swap": 0},
		"disk":   map[string]any{"/var/lib/snapd/snaps": map[string]any{"total": 100, "avail": 50}},
		"PCIdevices": []pci.PCIDevice{
			{Bus: "pci", Slot: "0000:00:02.0", BusNumber: 0x00, DeviceClass: 0x0300, VendorId: 0x8086, DeviceId: 0x5916},
		},
		"USBdevices": []usb.USBDevice{
			{Bus: "usb", BusNumber: 1, DeviceNumber: 2, VendorId: 0x0bda, ProductId: 0x5487},
		},
	}

	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	info, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	if _, ok := interface{}(info.PCIDevices[0]).(pci.PCIDevice); !ok {
		t.Fatalf("PCIDevices[0] type = %T, want pci.PCIDevice", info.PCIDevices[0])
	}
	if _, ok := interface{}(info.USBDevices[0]).(usb.USBDevice); !ok {
		t.Fatalf("USBDevices[0] type = %T, want usb.USBDevice", info.USBDevices[0])
	}
}

func TestDecodeMachine_InvalidDevice(t *testing.T) {
	// Build JSON directly with an unknown bus — Decode must return an error.
	data := []byte(`{"devices":[{"bus":"unknown","vendor-id":1}]}`)
	if _, err := Decode(data); err == nil {
		t.Fatal("expected error for unknown bus, got nil")
	}
}

func TestDecodeMachine_MalformedJSON(t *testing.T) {
	_, err := Decode([]byte(`not valid json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}
