package device

import (
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/device/fastrpc"
	"github.com/canonical/lscompute/pkg/machine/device/pci"
	"github.com/canonical/lscompute/pkg/machine/device/usb"
	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestDevices_WithFakeHost(t *testing.T) {
	machineRoot := filepath.Join("..", "..", "..", "test_data", "machines", "xps13-9350", "machine-root")
	h := host.Fake(machineRoot)

	devices, _, err := Devices(h, false)
	if err != nil {
		t.Fatalf("Devices() failed: %v", err)
	}

	if len(devices) == 0 {
		t.Fatal("expected at least one device, got none")
	}

	validBuses := map[string]bool{
		pci.BusName:     true,
		usb.BusName:     true,
		fastrpc.BusName: true,
	}
	for _, dev := range devices {
		if !validBuses[dev.Bus] {
			t.Errorf("device has unexpected Bus value %q", dev.Bus)
		}
	}
}
