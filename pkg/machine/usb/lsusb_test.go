package usb

import (
	"testing"
)

// --- lookupUsbIds ---

func TestLookupUsbIds(t *testing.T) {
	t.Run("known vendor and product", func(t *testing.T) {
		// 1d6b Linux Foundation / 0003 3.0 root hub
		entry, err := lookupUsbIds(0x1d6b, 0x0003)
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
		// 046d Logitech, Inc. / ffff is not a real product
		entry, err := lookupUsbIds(0x046d, 0xffff)
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
		entry, err := lookupUsbIds(0xffff, 0xffff)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.VendorName != "" || entry.ProductName != "" {
			t.Errorf("expected empty names for unknown ids, got vendor=%q product=%q", entry.VendorName, entry.ProductName)
		}
	})
}
