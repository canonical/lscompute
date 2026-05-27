package machine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestDevices_WithFakeHost(t *testing.T) {
	machineRoot := filepath.Join("..", "..", "test_data", "machines", "xps13-9350", "machine-root")
	if _, err := os.Stat(filepath.Join(machineRoot, "sys", "bus", "pci", "devices")); os.IsNotExist(err) {
		t.Skipf("fixture not present yet: %s", machineRoot)
	}
	h := host.Fake(machineRoot)

	devices, _, err := Devices(h, false)
	if err != nil {
		t.Fatalf("Devices() failed: %v", err)
	}

	if len(devices) == 0 {
		t.Fatal("expected at least one device, got none")
	}

	validBuses := map[string]bool{
		constants.BusPci:     true,
		constants.BusUsb:     true,
		constants.BusFastRpc: true,
	}
	for _, dev := range devices {
		if !validBuses[dev.Bus] {
			t.Errorf("device has unexpected Bus value %q", dev.Bus)
		}
	}
}
