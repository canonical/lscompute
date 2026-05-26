package machine

import (
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestGet_WithFakeHost(t *testing.T) {
	machineRoot := filepath.Join("..", "..", "test_data", "machines", "xps13-9350", "machine-root")
	h := host.Fake(machineRoot)

	info, _, err := Get(h, false)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if len(info.Cpus) == 0 {
		t.Error("expected at least one CPU, got none")
	}

	if info.Memory.TotalRam == 0 {
		t.Error("expected TotalRam > 0, got 0")
	}

	validBuses := map[string]bool{
		constants.BusPci:     true,
		constants.BusUsb:     true,
		constants.BusFastRpc: true,
	}
	for _, dev := range info.Devices {
		if !validBuses[dev.Bus] {
			t.Errorf("device has unexpected Bus value %q", dev.Bus)
		}
	}
}
