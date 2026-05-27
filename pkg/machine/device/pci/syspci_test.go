package pci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// writePciDevice writes the standard sysfs attribute files for a single PCI
// device slot under dir/sys/bus/pci/devices/<slot>/ and returns the host.
func writePciDevice(t *testing.T, dir, slot, vendor, device, class, subVendor, subDevice string) {
	t.Helper()
	slotDir := filepath.Join(dir, "sys", "bus", "pci", "devices", slot)
	if err := os.MkdirAll(slotDir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name, value string) {
		if err := os.WriteFile(filepath.Join(slotDir, name), []byte(value+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("vendor", vendor)
	write("device", device)
	write("class", class)
	if subVendor != "" {
		write("subsystem_vendor", subVendor)
	}
	if subDevice != "" {
		write("subsystem_device", subDevice)
	}
}

func TestReadHexFSFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    uint64
		wantErr bool
	}{
		{
			name:    "0x-prefixed hex value",
			content: "0x8086\n",
			want:    0x8086,
		},
		{
			name:    "plain hex value without prefix",
			content: "8086\n",
			want:    0x8086,
		},
		{
			name:    "value with whitespace",
			content: "  0x10de  \n",
			want:    0x10de,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			filePath := filepath.Join(dir, "value")
			if err := os.WriteFile(filePath, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}
			h := host.Fake(dir)
			got, err := readHexFSFile(h, "value")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got 0x%x, want 0x%x", got, tc.want)
			}
		})
	}

	t.Run("missing file returns error", func(t *testing.T) {
		h := host.Fake(t.TempDir())
		_, err := readHexFSFile(h, "nonexistent")
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}

func TestReadSysPciDevice(t *testing.T) {
	t.Run("full device with subsystem and zero prog-if", func(t *testing.T) {
		dir := t.TempDir()
		slot := "0000:3b:00.0"
		// class 0x030000: display controller, prog-if 0x00 → should be nil
		writePciDevice(t, dir, slot, "0x8086", "0x1234", "0x030000", "0x8086", "0x5678")
		h := host.Fake(dir)

		dev, err := readSysPciDevice(h, filepath.Join("sys", "bus", "pci", "devices", slot), slot)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.Slot != slot {
			t.Errorf("Slot: got %q, want %q", dev.Slot, slot)
		}
		if uint64(dev.BusNumber) != 0x3b {
			t.Errorf("BusNumber: got 0x%x, want 0x3b", uint64(dev.BusNumber))
		}
		if uint64(dev.VendorId) != 0x8086 {
			t.Errorf("VendorId: got 0x%x, want 0x8086", uint64(dev.VendorId))
		}
		if uint64(dev.DeviceId) != 0x1234 {
			t.Errorf("DeviceId: got 0x%x, want 0x1234", uint64(dev.DeviceId))
		}
		// DeviceClass is the upper 16 bits of the 24-bit class value (prog-if stripped)
		if uint64(dev.DeviceClass) != 0x0300 {
			t.Errorf("DeviceClass: got 0x%x, want 0x0300", uint64(dev.DeviceClass))
		}
		if dev.ProgrammingInterface != nil {
			t.Errorf("ProgrammingInterface: expected nil for prog-if 0x00, got %v", *dev.ProgrammingInterface)
		}
		if dev.SubvendorId == nil || uint64(*dev.SubvendorId) != 0x8086 {
			t.Errorf("SubvendorId: expected 0x8086, got %v", dev.SubvendorId)
		}
		if dev.SubdeviceId == nil || uint64(*dev.SubdeviceId) != 0x5678 {
			t.Errorf("SubdeviceId: expected 0x5678, got %v", dev.SubdeviceId)
		}
	})

	t.Run("non-zero prog-if sets ProgrammingInterface", func(t *testing.T) {
		dir := t.TempDir()
		slot := "0000:00:1f.2"
		// class 0x010601: SATA AHCI — prog-if 0x01
		writePciDevice(t, dir, slot, "0x8086", "0xa102", "0x010601", "", "")
		h := host.Fake(dir)

		dev, err := readSysPciDevice(h, filepath.Join("sys", "bus", "pci", "devices", slot), slot)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.ProgrammingInterface == nil {
			t.Fatal("ProgrammingInterface: expected non-nil for prog-if 0x01, got nil")
		}
		if *dev.ProgrammingInterface != 0x01 {
			t.Errorf("ProgrammingInterface: got 0x%02x, want 0x01", *dev.ProgrammingInterface)
		}
	})

	t.Run("missing subsystem files leaves SubvendorId and SubdeviceId nil", func(t *testing.T) {
		dir := t.TempDir()
		slot := "0000:00:02.0"
		writePciDevice(t, dir, slot, "0x8086", "0x9bc4", "0x030000", "", "")
		h := host.Fake(dir)

		dev, err := readSysPciDevice(h, filepath.Join("sys", "bus", "pci", "devices", slot), slot)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dev.SubvendorId != nil {
			t.Errorf("SubvendorId: expected nil, got 0x%x", uint64(*dev.SubvendorId))
		}
		if dev.SubdeviceId != nil {
			t.Errorf("SubdeviceId: expected nil, got 0x%x", uint64(*dev.SubdeviceId))
		}
	})

	t.Run("malformed slot returns error", func(t *testing.T) {
		dir := t.TempDir()
		slot := "badslot"
		slotDir := filepath.Join(dir, "sys", "bus", "pci", "devices", slot)
		if err := os.MkdirAll(slotDir, 0755); err != nil {
			t.Fatal(err)
		}
		h := host.Fake(dir)
		_, err := readSysPciDevice(h, filepath.Join("sys", "bus", "pci", "devices", slot), slot)
		if err == nil {
			t.Fatal("expected error for malformed slot, got nil")
		}
	})
}

func TestReadSysPci(t *testing.T) {
	t.Run("two valid devices are returned", func(t *testing.T) {
		dir := t.TempDir()
		writePciDevice(t, dir, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "0x8086", "0x0001")
		writePciDevice(t, dir, "0000:01:00.0", "0x10de", "0x2204", "0x030200", "0x1458", "0x4024")
		h := host.Fake(dir)

		devices, warnings, err := readSysPci(h)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings, got: %v", warnings)
		}
		if len(devices) != 2 {
			t.Fatalf("expected 2 devices, got %d", len(devices))
		}
	})

	t.Run("malformed slot produces warning and is skipped", func(t *testing.T) {
		dir := t.TempDir()
		// valid device
		writePciDevice(t, dir, "0000:00:02.0", "0x8086", "0x1234", "0x030000", "", "")
		// malformed slot directory (no colons)
		badSlotDir := filepath.Join(dir, "sys", "bus", "pci", "devices", "badslot")
		if err := os.MkdirAll(badSlotDir, 0755); err != nil {
			t.Fatal(err)
		}
		h := host.Fake(dir)

		devices, warnings, err := readSysPci(h)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(warnings) == 0 {
			t.Error("expected at least one warning for the malformed slot, got none")
		}
		if len(devices) != 1 {
			t.Errorf("expected 1 valid device, got %d", len(devices))
		}
	})
}
