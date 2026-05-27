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

// writeUsbIdsAt writes the given content to the given relative path inside a
// new temp directory and returns a fake host rooted there.
func writeUsbIdsAt(t *testing.T, relPath, content string) host.Host {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return host.Fake(dir)
}

// TestLookupUsbIds_NoDatabase verifies that an error is returned when none of
// the usb.ids search paths exist on the host.
func TestLookupUsbIds_NoDatabase(t *testing.T) {
	h := host.Fake(t.TempDir()) // empty dir — no usb.ids anywhere
	_, err := lookupUsbIds(h, 0x046d, 0xc52b)
	if err == nil {
		t.Fatal("expected error for missing database, got nil")
	}
}

// TestFindUsbIdsFile_AlternateSearchPaths verifies that usb.ids is found even
// when it lives in a secondary or tertiary candidate directory.
func TestFindUsbIdsFile_AlternateSearchPaths(t *testing.T) {
	paths := []struct {
		name    string
		relPath string
	}{
		{"usr/share/hwdata/usb.ids", "usr/share/hwdata/usb.ids"},
		{"usr/share/usb.ids", "usr/share/usb.ids"},
	}
	for _, p := range paths {
		p := p
		t.Run(p.name, func(t *testing.T) {
			h := writeUsbIdsAt(t, p.relPath, usbIdsFixture)
			// findUsbIdsFile must succeed and return the right path.
			got, err := findUsbIdsFile(h)
			if err != nil {
				t.Fatalf("findUsbIdsFile error: %v", err)
			}
			if got != p.relPath {
				t.Errorf("got path %q, want %q", got, p.relPath)
			}
			// Full lookup must also work through the alternate path.
			entry, err := lookupUsbIds(h, 0x1d6b, 0x0002)
			if err != nil {
				t.Fatalf("lookupUsbIds error: %v", err)
			}
			if entry.VendorName != "Linux Foundation" {
				t.Errorf("VendorName = %q, want %q", entry.VendorName, "Linux Foundation")
			}
		})
	}
}

// TestLookupUsbIds_ProductAtEOF verifies that a vendor+product pair found at
// the very end of the file (no trailing newline, vendor block closes at EOF)
// is resolved correctly.
func TestLookupUsbIds_ProductAtEOF(t *testing.T) {
	// "1d6b 0003" is the very last record — no blank line after it.
	fixture := "1d6b  Linux Foundation\n\t0003  3.0 root hub"
	h := writeUsbIdsAt(t, "usr/share/misc/usb.ids", fixture)

	entry, err := lookupUsbIds(h, 0x1d6b, 0x0003)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.VendorName != "Linux Foundation" {
		t.Errorf("VendorName = %q, want %q", entry.VendorName, "Linux Foundation")
	}
	if entry.ProductName != "3.0 root hub" {
		t.Errorf("ProductName = %q, want %q", entry.ProductName, "3.0 root hub")
	}
}

// TestLookupUsbIds_MalformedLines verifies that lines without the double-space
// separator are silently skipped and do not prevent correct entries from being
// found.
func TestLookupUsbIds_MalformedLines(t *testing.T) {
	fixture := "# comment\n" +
		"this-line-has-no-double-space\n" +
		"1d6b  Linux Foundation\n" +
		"\tbadline\n" + // product line missing double-space — skipped
		"\t0002  2.0 root hub\n"
	h := writeUsbIdsAt(t, "usr/share/misc/usb.ids", fixture)

	entry, err := lookupUsbIds(h, 0x1d6b, 0x0002)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.VendorName != "Linux Foundation" {
		t.Errorf("VendorName = %q, want %q", entry.VendorName, "Linux Foundation")
	}
	if entry.ProductName != "2.0 root hub" {
		t.Errorf("ProductName = %q, want %q", entry.ProductName, "2.0 root hub")
	}
}

// TestLookupUsbIds_InterfaceEntriesSkipped verifies that double-tab interface
// lines inside a vendor block do not interfere with product resolution.
func TestLookupUsbIds_InterfaceEntriesSkipped(t *testing.T) {
	fixture := "1d6b  Linux Foundation\n" +
		"\t0002  2.0 root hub\n" +
		"\t\t00  Full speed (or root) Hub\n" + // interface class — must be skipped
		"\t0003  3.0 root hub\n"
	h := writeUsbIdsAt(t, "usr/share/misc/usb.ids", fixture)

	entry, err := lookupUsbIds(h, 0x1d6b, 0x0003)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ProductName != "3.0 root hub" {
		t.Errorf("ProductName = %q, want %q", entry.ProductName, "3.0 root hub")
	}
}

// TestLookupFriendlyNames_LookupError verifies that lookupFriendlyNames returns
// the original device unchanged plus a non-nil error when the usb.ids database
// is missing.
func TestLookupFriendlyNames_LookupError(t *testing.T) {
	h := host.Fake(t.TempDir()) // no usb.ids
	dev := Device{}
	dev.VendorId = 0x046d
	dev.ProductId = 0xc52b

	got, err := lookupFriendlyNames(h, dev)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The device should be returned unchanged.
	if got.VendorName != nil || got.ProductName != nil {
		t.Errorf("expected unchanged device on error, got VendorName=%v ProductName=%v",
			got.VendorName, got.ProductName)
	}
}
