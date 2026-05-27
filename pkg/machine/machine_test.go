package machine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestGet_WithFakeHost(t *testing.T) {
	machineRoot := filepath.Join("..", "..", "test_data", "machines", "xps13-9350", "machine-root")
	if _, err := os.Stat(filepath.Join(machineRoot, "proc", "meminfo")); os.IsNotExist(err) {
		t.Skipf("fixture not present yet: %s", machineRoot)
	}
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

// TestGet_MemoryError verifies that Get returns an error when proc/meminfo is missing.
func TestGet_MemoryError(t *testing.T) {
	// Empty host — no proc/meminfo → memory.Info fails → Get must return an error.
	h := host.Fake(t.TempDir())
	_, _, err := Get(h, false)
	if err == nil {
		t.Fatal("expected error when proc/meminfo is missing, got nil")
	}
}

// TestGet_DevicesError verifies that Get propagates a device scan error.
// It builds a minimal fake host that satisfies memory/cpu/disk but breaks the PCI scanner.
func TestGet_DevicesError(t *testing.T) {
	// Inline minimal fixtures that let memory, cpu, and disk succeed.
	root := t.TempDir()

	write := func(rel, content string) {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// proc/meminfo — satisfies memory.Info
	write("proc/meminfo", "MemTotal: 8192000 kB\nSwapTotal: 0 kB\nSwapFree: 0 kB\n")
	// proc/sys/kernel/arch — satisfies machineArch (avoids uname syscall fallback)
	write("proc/sys/kernel/arch", "x86_64\n")
	// proc/cpuinfo — minimal amd64 entry satisfies cpu.Info
	write("proc/cpuinfo", "processor\t: 0\nvendor_id\t: GenuineIntel\nflags\t\t: sse\n")
	// run/disk-stats.json — satisfies disk.Info
	write("run/disk-stats.json",
		`{"/var/lib/snapd/snaps": {"total": 100000000000, "avail": 50000000000}}`)
	// sys/bus/pci/devices as a file (not directory) → ReadDir fails with a real error
	write("sys/bus/pci/devices", "not-a-dir")

	h := host.Fake(root)
	_, _, err := Get(h, false)
	if err == nil {
		t.Fatal("expected error when PCI devices path is a file, got nil")
	}
}
