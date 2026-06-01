package pci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestScannerBusName(t *testing.T) {
	s := NewScanner(Options{})
	if got := s.BusName(); got != BusName {
		t.Errorf("BusName() = %q, want %q", got, BusName)
	}
}

func TestScannerScan_EmptyHost(t *testing.T) {
	// Create an empty devices directory — scan must return 0 devices without error.
	root := t.TempDir()
	devicesDir := filepath.Join(root, "sys", "bus", "pci", "devices")
	if err := os.MkdirAll(devicesDir, 0755); err != nil {
		t.Fatal(err)
	}

	s := NewScanner(Options{})
	result, warnings, err := s.Scan(host.Fake(root))
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

	s := NewScanner(Options{})
	_, _, err := s.Scan(host.Fake(root))
	if err == nil {
		t.Fatal("expected Scan to return an error when devices path is a file, got nil")
	}
}

func TestScannerScan_NoFriendlyNames(t *testing.T) {
	root := t.TempDir()
	writePciDevice(t, root, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "", "")
	writePciDevice(t, root, "0000:01:00.0", "0x10de", "0x2204", "0x030200", "", "")

	s := NewScanner(Options{FriendlyNames: false})
	result, warnings, err := s.Scan(host.Fake(root))
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
		if di.Bus != BusName {
			t.Errorf("DeviceInfo.Bus = %q, want %q", di.Bus, BusName)
		}
		dev, ok := di.Payload.(*Device)
		if !ok {
			t.Fatalf("Payload is not *Device: %T", di.Payload)
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

	s := NewScanner(Options{FriendlyNames: true})
	result, warnings, err := s.Scan(host.Fake(root))
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
		dev, ok := di.Payload.(*Device)
		if !ok {
			t.Fatalf("Payload is not *Device: %T", di.Payload)
		}
		if dev.VendorName != nil {
			t.Errorf("expected nil VendorName on lookup failure, got %q", *dev.VendorName)
		}
	}
}

func TestScannerScan_PayloadImplementsBusDevice(t *testing.T) {
	root := t.TempDir()
	writePciDevice(t, root, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "0x8086", "0x0001")

	s := NewScanner(Options{})
	result, _, err := s.Scan(host.Fake(root))
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, di := range result {
		if di.Payload.BusName() != BusName {
			t.Errorf("Payload.BusName() = %q, want %q", di.Payload.BusName(), BusName)
		}
	}
}
