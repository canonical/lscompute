package usb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/device/bus"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func TestScannerScan_NoFriendlyNames(t *testing.T) {
	h := xps13Host(t)
	s := NewScanner(Options{FriendlyNames: false})

	result, warnings, err := s.Scan(h)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, w := range warnings {
		t.Logf("warning: %s", w)
	}

	const wantCount = 13
	if len(result) != wantCount {
		t.Errorf("expected %d DeviceInfo entries, got %d", wantCount, len(result))
	}

	for _, di := range result {
		if di.Bus != bus.BusUsb {
			t.Errorf("DeviceInfo.Bus = %q, want %q", di.Bus, bus.BusUsb)
		}
		dev, ok := di.Payload.(*Device)
		if !ok {
			t.Fatalf("Payload is not *Device: %T", di.Payload)
		}
		// No friendly names requested — names must be absent.
		if dev.VendorName != nil {
			t.Errorf("device %04x:%04x: expected nil VendorName without FriendlyNames, got %q",
				uint64(dev.VendorId), uint64(dev.ProductId), *dev.VendorName)
		}
		if dev.ProductName != nil {
			t.Errorf("device %04x:%04x: expected nil ProductName without FriendlyNames, got %q",
				uint64(dev.VendorId), uint64(dev.ProductId), *dev.ProductName)
		}
	}
}

func TestScannerScan_WithFriendlyNames(t *testing.T) {
	h := xps13Host(t)
	s := NewScanner(Options{FriendlyNames: true})

	result, warnings, err := s.Scan(h)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, w := range warnings {
		t.Logf("warning: %s", w)
	}

	const wantCount = 13
	if len(result) != wantCount {
		t.Errorf("expected %d DeviceInfo entries, got %d", wantCount, len(result))
	}

	// Build a lookup map for targeted assertions.
	type key struct{ vendorId, productId types.HexInt }
	byIds := map[key]*Device{}
	for i := range result {
		dev, ok := result[i].Payload.(*Device)
		if !ok {
			t.Fatalf("Payload is not *Device: %T", result[i].Payload)
		}
		byIds[key{dev.VendorId, dev.ProductId}] = dev
	}

	// knownBoth: vendor and product name both resolved from the curated usb.ids.
	knownBoth := []struct {
		vendorId    types.HexInt
		productId   types.HexInt
		wantVendor  string
		wantProduct string
	}{
		{0x1d6b, 0x0002, "Linux Foundation", "2.0 root hub"},
		{0x1d6b, 0x0003, "Linux Foundation", "3.0 root hub"},
		{0x046d, 0xc52b, "Logitech, Inc.", "Unifying Receiver"},
		{0x046d, 0xc548, "Logitech, Inc.", "Logi Bolt Receiver"},
		{0x0bda, 0x8153, "Realtek Semiconductor Corp.", "RTL8153 Gigabit Ethernet Adapter"},
	}
	cases := knownBoth

	for _, tc := range cases {
		k := key{tc.vendorId, tc.productId}
		dev, ok := byIds[k]
		if !ok {
			t.Errorf("device %04x:%04x not found in scan result", uint64(tc.vendorId), uint64(tc.productId))
			continue
		}
		if dev.VendorName == nil || *dev.VendorName != tc.wantVendor {
			got := "<nil>"
			if dev.VendorName != nil {
				got = *dev.VendorName
			}
			t.Errorf("device %04x:%04x VendorName = %q, want %q",
				uint64(tc.vendorId), uint64(tc.productId), got, tc.wantVendor)
		}
		if dev.ProductName == nil || *dev.ProductName != tc.wantProduct {
			got := "<nil>"
			if dev.ProductName != nil {
				got = *dev.ProductName
			}
			t.Errorf("device %04x:%04x ProductName = %q, want %q",
				uint64(tc.vendorId), uint64(tc.productId), got, tc.wantProduct)
		}
	}

	// knownVendorOnly: product ID not in the curated usb.ids, so only vendor is resolved.
	knownVendorOnly := []struct {
		vendorId   types.HexInt
		productId  types.HexInt
		wantVendor string
	}{
		{0x045e, 0x0840, "Microsoft Corp."},
		{0x27c6, 0x633c, "Shenzhen Goodix Technology Co.,Ltd."},
		{0x0bda, 0x5483, "Realtek Semiconductor Corp."},
		{0x0bda, 0x1100, "Realtek Semiconductor Corp."},
	}
	for _, tc := range knownVendorOnly {
		k := key{tc.vendorId, tc.productId}
		dev, ok := byIds[k]
		if !ok {
			t.Errorf("device %04x:%04x not found in scan result", uint64(tc.vendorId), uint64(tc.productId))
			continue
		}
		if dev.VendorName == nil || *dev.VendorName != tc.wantVendor {
			got := "<nil>"
			if dev.VendorName != nil {
				got = *dev.VendorName
			}
			t.Errorf("device %04x:%04x VendorName = %q, want %q",
				uint64(tc.vendorId), uint64(tc.productId), got, tc.wantVendor)
		}
		if dev.ProductName != nil {
			t.Errorf("device %04x:%04x: expected nil ProductName (not in db), got %q",
				uint64(tc.vendorId), uint64(tc.productId), *dev.ProductName)
		}
	}

	// Unknown vendor (2ac1) — no names should be populated even with FriendlyNames on.
	if dev, ok := byIds[key{0x2ac1, 0x20c9}]; ok {
		if dev.VendorName != nil {
			t.Errorf("unknown vendor 2ac1: expected nil VendorName, got %q", *dev.VendorName)
		}
		if dev.ProductName != nil {
			t.Errorf("unknown vendor 2ac1: expected nil ProductName, got %q", *dev.ProductName)
		}
	} else {
		t.Error("device 2ac1:20c9 not found in scan result")
	}
}

// TestScannerBusName verifies that the Scanner reports the canonical USB bus name.
func TestScannerBusName(t *testing.T) {
	s := NewScanner(Options{})
	if got := s.BusName(); got != bus.BusUsb {
		t.Errorf("BusName() = %q, want %q", got, bus.BusUsb)
	}
}


// TestScannerScan_SysFsError verifies that Scan returns a non-nil error when
// the sysfs USB devices path exists as a regular file (making ReadDir fail).
func TestScannerScan_SysFsError(t *testing.T) {
	root := t.TempDir()
	// Place a regular file where the devices directory should be.
	parent := filepath.Join(root, "sys", "bus", "usb")
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

// TestScannerScan_FriendlyNamesWarning verifies that when FriendlyNames is
// requested but no usb.ids database is available, Scan still returns all
// devices and emits one warning per device (no hard error).
func TestScannerScan_FriendlyNamesWarning(t *testing.T) {
	root := t.TempDir()

	// Two valid USB devices, but no usb.ids anywhere.
	makeUsbDeviceDir(t, root, "usb1", "1d6b", "0002", "1", "1")
	makeUsbDeviceDir(t, root, "usb2", "1d6b", "0003", "2", "1")

	s := NewScanner(Options{FriendlyNames: true})
	result, warnings, err := s.Scan(host.Fake(root))
	if err != nil {
		t.Fatalf("Scan() returned unexpected error: %v", err)
	}

	const wantDevices = 2
	if len(result) != wantDevices {
		t.Errorf("expected %d devices, got %d", wantDevices, len(result))
	}

	// Each device lookup should generate exactly one warning.
	if len(warnings) != wantDevices {
		t.Errorf("expected %d warnings (one per device), got %d: %v", wantDevices, len(warnings), warnings)
	}

	// Devices must still have no friendly names (lookup failed).
	for _, di := range result {
		dev, ok := di.Payload.(*Device)
		if !ok {
			t.Fatalf("Payload is not *Device: %T", di.Payload)
		}
		if dev.VendorName != nil {
			t.Errorf("expected nil VendorName on lookup failure, got %q", *dev.VendorName)
		}
		if dev.ProductName != nil {
			t.Errorf("expected nil ProductName on lookup failure, got %q", *dev.ProductName)
		}
	}

	// DeviceInfo.Bus must still be set correctly.
	for _, di := range result {
		if di.Bus != bus.BusUsb {
			t.Errorf("DeviceInfo.Bus = %q, want %q", di.Bus, bus.BusUsb)
		}
	}
}

// TestScannerScan_EmptyHost verifies that Scan on a host with no USB devices
// dir returns an empty slice without error.
func TestScannerScan_EmptyHost(t *testing.T) {
	s := NewScanner(Options{})
	result, warnings, err := s.Scan(host.Fake(t.TempDir()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

// Ensure the DeviceInfo payloads from Scan implement bus.BusDevice.
func TestScannerScan_PayloadImplementsBusDevice(t *testing.T) {
	h := xps13Host(t)
	s := NewScanner(Options{})
	result, _, err := s.Scan(h)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	for _, di := range result {
		bd, ok := di.Payload.(bus.BusDevice)
		if !ok {
			t.Errorf("Payload %T does not implement bus.BusDevice", di.Payload)
		} else if got := bd.BusName(); got != bus.BusUsb {
			t.Errorf("BusDevice.BusName() = %q, want %q", got, bus.BusUsb)
		}
	}
}
