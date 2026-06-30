package pci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func TestScannerScan_EmptyHost(t *testing.T) {
	// Create an empty devices directory — scan must return 0 devices without error.
	root := t.TempDir()
	devicesDir := filepath.Join(root, "sys", "bus", "pci", "devices")
	if err := os.MkdirAll(devicesDir, 0755); err != nil {
		t.Fatal(err)
	}

	bus := NewBus(host.Fake(root), Options{})
	result, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 devices, got %d", len(result))
	}
}

func TestScannerScan_SysFsError(t *testing.T) {
	// Place a regular file where the devices directory should be so ReadDir fails.
	root := t.TempDir()
	parent := filepath.Join(root, "sys", "bus", "pci")
	if err := os.MkdirAll(parent, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "devices"), []byte("not-a-dir"), 0644); err != nil {
		t.Fatal(err)
	}

	bus := NewBus(host.Fake(root), Options{})
	_, _, err := bus.Devices()
	if err == nil {
		t.Fatal("expected Scan to return an error when devices path is a file, got nil")
	}
}

func TestScannerScan_NoFriendlyNames(t *testing.T) {
	root := t.TempDir()
	writePciDevice(t, root, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "", "")
	writePciDevice(t, root, "0000:01:00.0", "0x10de", "0x2204", "0x030200", "", "")

	bus := NewBus(host.Fake(root), Options{FriendlyNames: false})
	result, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, w := range warnings {
		t.Logf("warning: %s", w)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 devices, got %d", len(result))
	}
	for _, di := range result {
		dev, ok := di.(Device)
		if !ok {
			t.Fatalf("item is not Device: %T", di)
		}
		if dev.Bus != BusName {
			t.Errorf("Device.Bus = %q, want %q", dev.Bus, BusName)
		}
		if dev.VendorName != nil {
			t.Errorf("expected nil VendorName without FriendlyNames, got %q", *dev.VendorName)
		}
	}
}

func TestScannerScan_FriendlyNamesWarning(t *testing.T) {
	// Valid PCI devices but no pci.ids database → warning per device, no error.
	root := t.TempDir()
	writePciDevice(t, root, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "", "")
	writePciDevice(t, root, "0000:01:00.0", "0x10de", "0x2204", "0x030200", "", "")

	bus := NewBus(host.Fake(root), Options{FriendlyNames: true})
	result, warnings, err := bus.Devices()
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 devices, got %d", len(result))
	}
	// Each device should generate a warning (no pci.ids).
	if len(warnings) == 0 {
		t.Error("expected warnings for missing pci.ids, got none")
	}
	// No friendly names should have been populated.
	for _, di := range result {
		dev, ok := di.(Device)
		if !ok {
			t.Fatalf("item is not Device: %T", di)
		}
		if dev.VendorName != nil {
			t.Errorf("expected nil VendorName on lookup failure, got %q", *dev.VendorName)
		}
	}
}

func TestScannerScan_FriendlyNamesSuccess(t *testing.T) {
	// Write a PCI device and a pci.ids database so that lookupFriendlyNames succeeds.
	root := t.TempDir()
	writePciDevice(t, root, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "", "")

	misc := filepath.Join(root, "usr", "share", "misc")
	if err := os.MkdirAll(misc, 0755); err != nil {
		t.Fatal(err)
	}
	pciIds := "8086  Intel Corporation\n\t1234  Fake Display Device\n"
	if err := os.WriteFile(filepath.Join(misc, "pci.ids"), []byte(pciIds), 0644); err != nil {
		t.Fatal(err)
	}

	bus := NewBus(host.Fake(root), Options{FriendlyNames: true})
	result, _, err := bus.Devices()
	if err != nil {
		t.Fatalf("Devices() unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 device, got %d", len(result))
	}
	dev, ok := result[0].(Device)
	if !ok {
		t.Fatalf("item is not Device: %T", result[0])
	}
	if dev.VendorName == nil || *dev.VendorName != "Intel Corporation" {
		t.Errorf("VendorName = %v, want %q", dev.VendorName, "Intel Corporation")
	}
	if dev.DeviceName == nil || *dev.DeviceName != "Fake Display Device" {
		t.Errorf("DeviceName = %v, want %q", dev.DeviceName, "Fake Display Device")
	}
}

func TestIsGpu(t *testing.T) {
	cases := []struct {
		name        string
		deviceClass uint64
		want        bool
	}{
		{"VGA legacy (0x0001)", 0x0001, true},
		{"display controller (0x0300)", 0x0300, true},
		{"3D controller (0x0302)", 0x0302, true},
		{"display — any 0x03xx subclass", 0x0301, true},
		{"network controller (0x0200)", 0x0200, false},
		{"storage controller (0x0100)", 0x0100, false},
		{"audio device (0x0403)", 0x0403, false},
		{"USB host controller (0x0c03)", 0x0c03, false},
		{"zero / unclassified (0x0000)", 0x0000, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := Device{DeviceClass: types.HexInt(tc.deviceClass)}
			if got := d.IsGpu(); got != tc.want {
				t.Errorf("Device{DeviceClass: 0x%04x}.IsGpu() = %v, want %v",
					tc.deviceClass, got, tc.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	t.Run("valid JSON round-trip", func(t *testing.T) {
		raw := `{
			"bus": "pci",
			"slot": "0000:00:02.0",
			"bus-number": "0x00",
			"device-class": "0x0300",
			"vendor-id": "0x8086",
			"device-id": "0x1234"
		}`
		dev, err := DecodeJSON([]byte(raw))
		if err != nil {
			t.Fatalf("DecodeJSON() unexpected error: %v", err)
		}
		if dev.Bus != "pci" {
			t.Errorf("Bus = %q, want %q", dev.Bus, "pci")
		}
		if dev.Slot != "0000:00:02.0" {
			t.Errorf("Slot = %q, want %q", dev.Slot, "0000:00:02.0")
		}
		if dev.VendorId != types.HexInt(0x8086) {
			t.Errorf("VendorId = 0x%x, want 0x8086", uint64(dev.VendorId))
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, err := DecodeJSON([]byte(`{not valid json`))
		if err == nil {
			t.Fatal("DecodeJSON() expected error for invalid JSON, got nil")
		}
	})
}
