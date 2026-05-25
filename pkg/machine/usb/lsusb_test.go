package usb

import (
	"testing"

	"github.com/canonical/lscompute/pkg/machine/types"
	"github.com/go-test/deep"
)

// --- parseLsUsbLine ---

func TestParseLsUsbLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    types.UsbDevice
		wantErr bool
	}{
		{
			name: "standard line with description",
			line: "Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub",
			want: types.UsbDevice{BusNumber: 2, DeviceNumber: 1, VendorId: 0x1d6b, ProductId: 0x0003},
		},
		{
			name: "line without description",
			line: "Bus 001 Device 001: ID 1d6b:0002",
			want: types.UsbDevice{BusNumber: 1, DeviceNumber: 1, VendorId: 0x1d6b, ProductId: 0x0002},
		},
		{
			name: "high product id",
			line: "Bus 003 Device 002: ID 046d:c52b Logitech, Inc. Unifying Receiver",
			want: types.UsbDevice{BusNumber: 3, DeviceNumber: 2, VendorId: 0x046d, ProductId: 0xc52b},
		},
		{
			name:    "missing ID marker",
			line:    "Bus 001 Device 001: 1d6b:0002",
			wantErr: true,
		},
		{
			name:    "malformed vendor:product",
			line:    "Bus 001 Device 001: ID 1d6b-0002",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseLsUsbLine(tc.line)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := deep.Equal(got, tc.want); diff != nil {
				t.Errorf("mismatch: %v", diff)
			}
		})
	}
}

// --- ParseLsUsb ---

func TestParseLsUsb(t *testing.T) {
	input := `Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub
Bus 001 Device 003: ID 1050:0407 Yubico.com Yubikey 4/5 OTP+U2F+CCID
Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
Bus 003 Device 002: ID 046d:c52b Logitech, Inc. Unifying Receiver
`

	t.Run("without friendly names - no names set", func(t *testing.T) {
		devices, _, err := ParseLsUsb(input, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(devices) != 4 {
			t.Fatalf("expected 4 devices, got %d", len(devices))
		}
		for _, d := range devices {
			if d.VendorName != nil {
				t.Errorf("expected no vendor name, got %q", *d.VendorName)
			}
			if d.ProductName != nil {
				t.Errorf("expected no product name, got %q", *d.ProductName)
			}
		}
	})

	t.Run("with friendly names - separate vendor and product", func(t *testing.T) {
		devices, warnings, err := ParseLsUsb(input, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(warnings) > 0 {
			t.Logf("warnings: %v", warnings)
		}
		if len(devices) != 4 {
			t.Fatalf("expected 4 devices, got %d", len(devices))
		}

		// Linux Foundation (0x1d6b) / 3.0 root hub (0x0003)
		d := devices[0]
		if d.VendorName == nil || *d.VendorName != "Linux Foundation" {
			t.Errorf("devices[0]: expected vendor-name 'Linux Foundation', got %v", d.VendorName)
		}
		if d.ProductName == nil || *d.ProductName != "3.0 root hub" {
			t.Errorf("devices[0]: expected product-name '3.0 root hub', got %v", d.ProductName)
		}

		// Logitech (0x046d) / Unifying Receiver (0xc52b)
		d = devices[3]
		if d.VendorName == nil || *d.VendorName != "Logitech, Inc." {
			t.Errorf("devices[3]: expected vendor-name 'Logitech, Inc.', got %v", d.VendorName)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		devices, warnings, err := ParseLsUsb("", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got %v", warnings)
		}
		if len(devices) != 0 {
			t.Errorf("expected no devices, got %d", len(devices))
		}
	})
}

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
