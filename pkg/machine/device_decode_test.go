package machine

import (
	"encoding/json"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/pci"
	"github.com/canonical/lscompute/pkg/machine/types"
	"github.com/canonical/lscompute/pkg/machine/usb"
)

func TestDecodeDeviceInfo(t *testing.T) {
	t.Run("decodes pci device", func(t *testing.T) {
		data := []byte(`{"bus":"pci","slot":"0000:00:02.0","bus-number":"0x00","device-class":"0x0300","vendor-id":"0x8086","device-id":"0x5916"}`)
		dev, err := DecodeDeviceInfo(data)
		if err != nil {
			t.Fatalf("DecodeDeviceInfo() error: %v", err)
		}
		if dev.Bus != constants.BusPci {
			t.Fatalf("Bus = %q, want %q", dev.Bus, constants.BusPci)
		}
		if _, ok := dev.Payload.(*pci.Device); !ok {
			t.Fatalf("Payload type = %T, want *pci.Device", dev.Payload)
		}
	})

	t.Run("decodes usb device", func(t *testing.T) {
		data := []byte(`{"bus":"usb","bus-number":1,"device-number":2,"vendor-id":"0x0bda","product-id":"0x5487"}`)
		dev, err := DecodeDeviceInfo(data)
		if err != nil {
			t.Fatalf("DecodeDeviceInfo() error: %v", err)
		}
		if dev.Bus != constants.BusUsb {
			t.Fatalf("Bus = %q, want %q", dev.Bus, constants.BusUsb)
		}
		if _, ok := dev.Payload.(*usb.Device); !ok {
			t.Fatalf("Payload type = %T, want *usb.Device", dev.Payload)
		}
	})

	t.Run("unknown bus returns error", func(t *testing.T) {
		_, err := DecodeDeviceInfo([]byte(`{"bus":"unknown","vendor-id":1}`))
		if err == nil {
			t.Fatal("expected error for unknown bus, got nil")
		}
	})
}

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

	info, err := DecodeMachineInfo(data)
	if err != nil {
		t.Fatalf("DecodeMachineInfo() error: %v", err)
	}
	if len(info.Devices) != 2 {
		t.Fatalf("len(Devices) = %d, want 2", len(info.Devices))
	}
	if _, ok := info.Devices[0].Payload.(*usb.Device); !ok {
		t.Fatalf("Devices[0] payload type = %T, want *usb.Device", info.Devices[0].Payload)
	}
	if _, ok := info.Devices[1].Payload.(*pci.Device); !ok {
		t.Fatalf("Devices[1] payload type = %T, want *pci.Device", info.Devices[1].Payload)
	}
}

func TestDecodeMachineInfo_InvalidDevice(t *testing.T) {
	info := types.MachineInfo{
		Devices: []types.DeviceInfo{{Bus: "unknown", Payload: &usb.Device{BusNumber: 1}}},
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	if _, err := DecodeMachineInfo(data); err == nil {
		t.Fatal("expected error for unknown bus, got nil")
	}
}
