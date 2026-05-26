package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// meminfoFixture is a minimal /proc/meminfo that parseProcMemInfo can handle.
const meminfoFixture = `MemTotal:       16326508 kB
MemFree:          354360 kB
MemAvailable:    8000000 kB
SwapTotal:       2097148 kB
SwapFree:        2097148 kB
`

func writeMeminfo(t *testing.T, content string) host.Host {
	t.Helper()
	dir := t.TempDir()
	procDir := filepath.Join(dir, "proc")
	if err := os.MkdirAll(procDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(procDir, "meminfo"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return host.Fake(dir)
}

func TestMemoryInfo(t *testing.T) {
	h := writeMeminfo(t, meminfoFixture)

	info, err := Info(h)
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}

	// 16326508 KiB → 16526344192 bytes
	const wantTotal = 16326508 * 1024
	if info.TotalRam != wantTotal {
		t.Errorf("TotalRam = %d, want %d", info.TotalRam, wantTotal)
	}
}

func TestMemoryInfo_MissingFile(t *testing.T) {
	// Empty host — no proc/meminfo present.
	h := host.Fake(t.TempDir())
	_, err := Info(h)
	if err == nil {
		t.Fatal("expected error for missing proc/meminfo, got nil")
	}
}
