package disk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func TestDiskInfo(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, "run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Fixture JSON uses leading "/" keys, matching the host.Fake convention.
	fixture := `{"/var/lib/snapd/snaps": {"total": 107374182400, "avail": 21474836480}}`
	if err := os.WriteFile(filepath.Join(runDir, "disk-stats.json"), []byte(fixture), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	result, err := Info(h)
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}

	stats, ok := result[snapStoragePath]
	if !ok {
		t.Fatalf("expected key %q in result, got keys: %v", snapStoragePath, keysOf(result))
	}
	if stats.Total != 107374182400 {
		t.Errorf("Total = %d, want 107374182400", stats.Total)
	}
	if stats.Avail != 21474836480 {
		t.Errorf("Avail = %d, want 21474836480", stats.Avail)
	}
}

func TestDiskInfo_MissingStats(t *testing.T) {
	// No run/disk-stats.json → StatFs should fail → Info returns an error.
	h := host.Fake(t.TempDir())
	_, err := Info(h)
	if err == nil {
		t.Fatal("expected error for missing disk-stats.json, got nil")
	}
}

func keysOf[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
