package pci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// pciIdsFixture is a curated subset of a real pci.ids database that exercises
// the parsing paths we care about.
const pciIdsFixture = `#
# Sample pci.ids
#
8086  Intel Corporation
	1234  Fake Display Device
		8086 5678  Intel Reference Subsystem
	abcd  Another Intel Device
10de  NVIDIA Corporation
	2204  GA102 [GeForce RTX 3090]
		1458 4024  Gigabyte GeForce RTX 3090
1002  Advanced Micro Devices, Inc. [AMD/ATI]
	687f  Vega 20
C 00  Unclassified device
`

func writePciIdsFixture(t *testing.T) host.Host {
	t.Helper()
	dir := t.TempDir()
	misc := filepath.Join(dir, "usr", "share", "misc")
	if err := os.MkdirAll(misc, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(misc, "pci.ids"), []byte(pciIdsFixture), 0644); err != nil {
		t.Fatal(err)
	}
	return host.Fake(dir)
}

func TestLookupPciIds(t *testing.T) {
	h := writePciIdsFixture(t)

	sv := types.HexInt(0x8086)
	sd := types.HexInt(0x5678)

	tests := []struct {
		name          string
		vendorId      types.HexInt
		deviceId      types.HexInt
		subvendorId   *types.HexInt
		subdeviceId   *types.HexInt
		wantVendor    string
		wantDevice    string
		wantSubvendor string
		wantSubdevice string
	}{
		{
			name:       "known vendor and device",
			vendorId:   0x8086,
			deviceId:   0x1234,
			wantVendor: "Intel Corporation",
			wantDevice: "Fake Display Device",
		},
		{
			name:          "known vendor, device, subvendor and subdevice",
			vendorId:      0x8086,
			deviceId:      0x1234,
			subvendorId:   &sv,
			subdeviceId:   &sd,
			wantVendor:    "Intel Corporation",
			wantDevice:    "Fake Display Device",
			wantSubvendor: "Intel Corporation",
			wantSubdevice: "Intel Reference Subsystem",
		},
		{
			name:       "known vendor with unknown device",
			vendorId:   0x10de,
			deviceId:   0xffff,
			wantVendor: "NVIDIA Corporation",
			wantDevice: "",
		},
		{
			name:       "fully unknown ids",
			vendorId:   0xffff,
			deviceId:   0xffff,
			wantVendor: "",
			wantDevice: "",
		},
		{
			name:       "second device under same vendor",
			vendorId:   0x8086,
			deviceId:   0xabcd,
			wantVendor: "Intel Corporation",
			wantDevice: "Another Intel Device",
		},
		{
			name:       "nvidia with known device",
			vendorId:   0x10de,
			deviceId:   0x2204,
			wantVendor: "NVIDIA Corporation",
			wantDevice: "GA102 [GeForce RTX 3090]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := lookupPciIds(h, tc.vendorId, tc.deviceId, tc.subvendorId, tc.subdeviceId)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.VendorName != tc.wantVendor {
				t.Errorf("VendorName: got %q, want %q", got.VendorName, tc.wantVendor)
			}
			if got.DeviceName != tc.wantDevice {
				t.Errorf("DeviceName: got %q, want %q", got.DeviceName, tc.wantDevice)
			}
			if got.SubvendorName != tc.wantSubvendor {
				t.Errorf("SubvendorName: got %q, want %q", got.SubvendorName, tc.wantSubvendor)
			}
			if got.SubdeviceName != tc.wantSubdevice {
				t.Errorf("SubdeviceName: got %q, want %q", got.SubdeviceName, tc.wantSubdevice)
			}
		})
	}
}

func TestLookupPciIds_MissingFile(t *testing.T) {
	h := host.Fake(t.TempDir()) // no pci.ids written
	_, err := lookupPciIds(h, 0x8086, 0x1234, nil, nil)
	if err == nil {
		t.Fatal("expected error when pci.ids is missing, got nil")
	}
}

func TestSplitPciIdName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantId   string
		wantName string
		wantOk   bool
	}{
		{
			name:     "well-formed vendor line",
			input:    "8086  Intel Corporation",
			wantId:   "8086",
			wantName: "Intel Corporation",
			wantOk:   true,
		},
		{
			name:     "well-formed device line (tab-stripped)",
			input:    "1234  Fake Display Device",
			wantId:   "1234",
			wantName: "Fake Display Device",
			wantOk:   true,
		},
		{
			name:     "name with extra spaces trimmed",
			input:    "abcd    Some Device With Spaces",
			wantId:   "abcd",
			wantName: "Some Device With Spaces",
			wantOk:   true,
		},
		{
			name:   "missing double-space separator",
			input:  "8086 Intel Corporation",
			wantOk: false,
		},
		{
			name:   "comment line",
			input:  "# this is a comment",
			wantOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, name, ok := splitPciIdName(tc.input)
			if ok != tc.wantOk {
				t.Fatalf("ok: got %v, want %v", ok, tc.wantOk)
			}
			if !ok {
				return
			}
			if id != tc.wantId {
				t.Errorf("id: got %q, want %q", id, tc.wantId)
			}
			if name != tc.wantName {
				t.Errorf("name: got %q, want %q", name, tc.wantName)
			}
		})
	}
}

func TestSplitSubsystemLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSV   string
		wantSD   string
		wantName string
		wantOk   bool
	}{
		{
			name:     "well-formed subsystem line",
			input:    "8086 5678  Intel Reference Subsystem",
			wantSV:   "8086",
			wantSD:   "5678",
			wantName: "Intel Reference Subsystem",
			wantOk:   true,
		},
		{
			name:     "gigabyte subsystem",
			input:    "1458 4024  Gigabyte GeForce RTX 3090",
			wantSV:   "1458",
			wantSD:   "4024",
			wantName: "Gigabyte GeForce RTX 3090",
			wantOk:   true,
		},
		{
			name:   "missing double-space before name",
			input:  "8086 5678 No double space here",
			wantOk: false,
		},
		{
			name:   "only one id field",
			input:  "8086  Only one ID",
			wantOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sv, sd, name, ok := splitSubsystemLine(tc.input)
			if ok != tc.wantOk {
				t.Fatalf("ok: got %v, want %v", ok, tc.wantOk)
			}
			if !ok {
				return
			}
			if sv != tc.wantSV {
				t.Errorf("subvendor: got %q, want %q", sv, tc.wantSV)
			}
			if sd != tc.wantSD {
				t.Errorf("subdevice: got %q, want %q", sd, tc.wantSD)
			}
			if name != tc.wantName {
				t.Errorf("name: got %q, want %q", name, tc.wantName)
			}
		})
	}
}

// TestLookupFriendlyNames exercises the higher-level wrapper that converts the
// raw pciIdEntry result into a FriendlyNames struct with optional string pointers.
func TestLookupFriendlyNames(t *testing.T) {
	h := writePciIdsFixture(t)

	sv := types.HexInt(0x8086)
	sd := types.HexInt(0x5678)

	t.Run("known vendor and device populate names", func(t *testing.T) {
		dev := Device{VendorId: 0x8086, DeviceId: 0x1234}
		names, err := lookupFriendlyNames(h, dev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if names.VendorName == nil || *names.VendorName != "Intel Corporation" {
			t.Errorf("VendorName = %v, want %q", names.VendorName, "Intel Corporation")
		}
		if names.DeviceName == nil || *names.DeviceName != "Fake Display Device" {
			t.Errorf("DeviceName = %v, want %q", names.DeviceName, "Fake Display Device")
		}
		if names.SubvendorName != nil {
			t.Errorf("SubvendorName should be nil when not requested, got %q", *names.SubvendorName)
		}
	})

	t.Run("known subsystem populates all four names", func(t *testing.T) {
		dev := Device{
			VendorId:    0x8086,
			DeviceId:    0x1234,
			SubvendorId: &sv,
			SubdeviceId: &sd,
		}
		names, err := lookupFriendlyNames(h, dev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if names.SubvendorName == nil || *names.SubvendorName != "Intel Corporation" {
			t.Errorf("SubvendorName = %v, want %q", names.SubvendorName, "Intel Corporation")
		}
		if names.SubdeviceName == nil || *names.SubdeviceName != "Intel Reference Subsystem" {
			t.Errorf("SubdeviceName = %v, want %q", names.SubdeviceName, "Intel Reference Subsystem")
		}
	})

	t.Run("unknown vendor returns all-nil FriendlyNames", func(t *testing.T) {
		dev := Device{VendorId: 0xffff, DeviceId: 0xffff}
		names, err := lookupFriendlyNames(h, dev)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if names.VendorName != nil || names.DeviceName != nil {
			t.Errorf("expected all-nil names for unknown vendor, got vendor=%v device=%v",
				names.VendorName, names.DeviceName)
		}
	})

	t.Run("missing pci.ids returns error", func(t *testing.T) {
		emptyHost := host.Fake(t.TempDir())
		dev := Device{VendorId: 0x8086, DeviceId: 0x1234}
		_, err := lookupFriendlyNames(emptyHost, dev)
		if err == nil {
			t.Fatal("expected error for missing pci.ids, got nil")
		}
	})
}
