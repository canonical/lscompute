package usb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// usbIdsFixture is a curated subset of the real usb.ids database that exercises
// the parsing paths we care about: known vendor + product, known vendor with
// unknown product, and unknown vendor.
const usbIdsFixture = `#
# Sample usb.ids
#
046d  Logitech, Inc.
	c52b  Unifying Receiver
	c534  Unifying Receiver
1d6b  Linux Foundation
	0001  1.1 root hub
	0002  2.0 root hub
	0003  3.0 root hub
`

func writeUsbIdsFixture(t *testing.T) host.Host {
	t.Helper()
	dir := t.TempDir()
	misc := filepath.Join(dir, "usr", "share", "misc")
	if err := os.MkdirAll(misc, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(misc, "usb.ids"), []byte(usbIdsFixture), 0644); err != nil {
		t.Fatal(err)
	}
	return host.Fake(dir)
}

func TestLookupUsbIds(t *testing.T) {
	h := writeUsbIdsFixture(t)

	t.Run("known vendor and product", func(t *testing.T) {
		entry, err := lookupUsbIds(h, 0x1d6b, 0x0003)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.VendorName != "Linux Foundation" {
			t.Errorf("expected vendor 'Linux Foundation', got %q", entry.VendorName)
		}
		if entry.ProductName != "3.0 root hub" {
			t.Errorf("expected product '3.0 root hub', got %q", entry.ProductName)
		}
	})

	t.Run("known vendor, unknown product", func(t *testing.T) {
		entry, err := lookupUsbIds(h, 0x046d, 0xffff)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.VendorName != "Logitech, Inc." {
			t.Errorf("expected vendor 'Logitech, Inc.', got %q", entry.VendorName)
		}
		if entry.ProductName != "" {
			t.Errorf("expected empty product name, got %q", entry.ProductName)
		}
	})

	t.Run("unknown vendor", func(t *testing.T) {
		entry, err := lookupUsbIds(h, 0xffff, 0xffff)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.VendorName != "" || entry.ProductName != "" {
			t.Errorf("expected empty names for unknown ids, got vendor=%q product=%q", entry.VendorName, entry.ProductName)
		}
	})
}
