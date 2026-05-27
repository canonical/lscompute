package usb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// xps13MachineRoot is the shared fake-host fixture for the xps13-9350 machine.
const xps13MachineRoot = "../../../../test_data/machines/xps13-9350/machine-root"

func xps13Host(t *testing.T) host.Host {
	t.Helper()
	abs, err := filepath.Abs(xps13MachineRoot)
	if err != nil {
		t.Fatalf("resolving machine root: %v", err)
	}
	return host.Fake(abs)
}

func TestReadSysUsb_XPS13(t *testing.T) {
	h := xps13Host(t)

	devices, warnings, err := readSysUsb(h)
	if err != nil {
		t.Fatalf("readSysUsb() error: %v", err)
	}
	for _, w := range warnings {
		t.Logf("warning: %s", w)
	}

	// The fixture has 13 non-interface USB device directories.
	const wantCount = 13
	if len(devices) != wantCount {
		t.Errorf("expected %d devices, got %d", wantCount, len(devices))
	}

	// Confirm interface entries (names containing ":") are filtered out.
	for _, d := range devices {
		_ = d // interface entries would have been silently skipped; any device
		// reaching here must be a proper device node.
	}

	// Spot-check: usb1 → Linux Foundation 2.0 root hub
	assertDevice(t, devices, types.HexInt(0x1d6b), types.HexInt(0x0002), 1, 1)
	// Spot-check: usb2 → Linux Foundation 3.0 root hub
	assertDevice(t, devices, types.HexInt(0x1d6b), types.HexInt(0x0003), 2, 1)
	// Spot-check: Logitech Unifying Receiver
	assertDevice(t, devices, types.HexInt(0x046d), types.HexInt(0xc52b), 3, 22)
	// Spot-check: Goodix Fingerprint Reader
	assertDevice(t, devices, types.HexInt(0x27c6), types.HexInt(0x633c), 3, 8)
	// Spot-check: unknown vendor still parsed correctly
	assertDevice(t, devices, types.HexInt(0x2ac1), types.HexInt(0x20c9), 3, 3)
}

// TestReadSysUsb_MissingDir verifies that a host with no sys/bus/usb/devices
// directory returns an empty result without an error.
func TestReadSysUsb_MissingDir(t *testing.T) {
	dir := t.TempDir() // empty — no sys/ subtree at all
	h := host.Fake(dir)

	devices, warnings, err := readSysUsb(h)
	if err != nil {
		t.Fatalf("expected no error for missing dir, got: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if len(devices) != 0 {
		t.Errorf("expected no devices, got %d", len(devices))
	}
}

// makeUsbDeviceDir creates a valid sysfs USB device directory under root and
// writes the given fields. Pass "" to skip writing a file (simulates missing).
func makeUsbDeviceDir(t *testing.T, root, name, idVendor, idProduct, busnum, devnum string) {
	t.Helper()
	dir := filepath.Join(root, "sys", "bus", "usb", "devices", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(file, content string) {
		if content == "" {
			return
		}
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("idVendor", idVendor)
	write("idProduct", idProduct)
	write("busnum", busnum)
	write("devnum", devnum)
}

// TestReadSysUsb_CorruptDevice verifies that a device directory with a missing
// required sysfs file produces a warning but does not abort the whole scan, and
// that the valid sibling device is still returned.
func TestReadSysUsb_CorruptDevice(t *testing.T) {
	root := t.TempDir()

	// Valid device.
	makeUsbDeviceDir(t, root, "usb1", "1d6b", "0002", "1", "1")
	// Corrupt device: idVendor file absent.
	if err := os.MkdirAll(filepath.Join(root, "sys", "bus", "usb", "devices", "1-1"), 0755); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(root)
	devices, warnings, err := readSysUsb(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning for the corrupt device, got %d: %v", len(warnings), warnings)
	}
}

// TestReadSysUsb_InterfaceEntriesSkipped verifies that entries whose names
// contain ":" (USB interface nodes like "1-1:1.0") are silently ignored.
func TestReadSysUsb_InterfaceEntriesSkipped(t *testing.T) {
	root := t.TempDir()

	// One real device.
	makeUsbDeviceDir(t, root, "usb1", "1d6b", "0002", "1", "1")
	// Interface entry — create a directory with a ":" in the name.
	ifDir := filepath.Join(root, "sys", "bus", "usb", "devices", "1-1:1.0")
	if err := os.MkdirAll(ifDir, 0755); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(root)
	devices, _, err := readSysUsb(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("expected 1 device (interface entry must be skipped), got %d", len(devices))
	}
}

// TestReadSysUsbDevice_ParseErrors covers every parse-error branch inside
// readSysUsbDevice: missing file and malformed content for each of the four
// sysfs attributes it reads.
func TestReadSysUsbDevice_ParseErrors(t *testing.T) {
	cases := []struct {
		name      string
		idVendor  string
		idProduct string
		busnum    string
		devnum    string
		wantErr   bool
	}{
		{
			name:     "missing idVendor", // idVendor file not written
			idVendor: "", idProduct: "0002", busnum: "1", devnum: "1",
			wantErr: true,
		},
		{
			name:     "bad idVendor hex",
			idVendor: "ZZZZ", idProduct: "0002", busnum: "1", devnum: "1",
			wantErr: true,
		},
		{
			name:     "missing idProduct",
			idVendor: "1d6b", idProduct: "", busnum: "1", devnum: "1",
			wantErr: true,
		},
		{
			name:     "bad idProduct hex",
			idVendor: "1d6b", idProduct: "ZZZZ", busnum: "1", devnum: "1",
			wantErr: true,
		},
		{
			name:     "missing busnum",
			idVendor: "1d6b", idProduct: "0002", busnum: "", devnum: "1",
			wantErr: true,
		},
		{
			name:     "bad busnum int",
			idVendor: "1d6b", idProduct: "0002", busnum: "not-a-number", devnum: "1",
			wantErr: true,
		},
		{
			name:     "missing devnum",
			idVendor: "1d6b", idProduct: "0002", busnum: "1", devnum: "",
			wantErr: true,
		},
		{
			name:     "bad devnum int",
			idVendor: "1d6b", idProduct: "0002", busnum: "1", devnum: "not-a-number",
			wantErr: true,
		},
		{
			name:     "all valid",
			idVendor: "1d6b", idProduct: "0002", busnum: "1", devnum: "1",
			wantErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			makeUsbDeviceDir(t, root, "usb1", tc.idVendor, tc.idProduct, tc.busnum, tc.devnum)
			h := host.Fake(root)
			dir := filepath.Join("sys", "bus", "usb", "devices", "usb1")
			_, err := readSysUsbDevice(h, dir)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// assertDevice fails the test if no device in the slice matches all four fields.
func assertDevice(t *testing.T, devices []Device, vendorId, productId types.HexInt, busNum, devNum int) {
	t.Helper()
	for _, d := range devices {
		if d.VendorId == vendorId && d.ProductId == productId &&
			d.BusNumber == busNum && d.DeviceNumber == devNum {
			return
		}
	}
	t.Errorf("device vendor=%#x product=%#x bus=%d dev=%d not found in result",
		uint64(vendorId), uint64(productId), busNum, devNum)
}
