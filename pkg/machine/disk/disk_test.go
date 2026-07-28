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
	// disk-stats.json is an array of {mountpoint, total, avail} objects.
	fixture := `[{"mountpoint": "/", "total": 107374182400, "avail": 21474836480}]`
	if err := os.WriteFile(filepath.Join(runDir, "disk-stats.json"), []byte(fixture), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	results, err := Info(h)
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}

	for _, result := range results {
		if result.MountPoint != "/" {
			t.Errorf("unexpected mount point: got %q, want %q", result.MountPoint, "/")
		}
		if result.Total != 107374182400 {
			t.Errorf("Total = %d, want 107374182400", result.Total)
		}
		if result.Available != 21474836480 {
			t.Errorf("Available = %d, want 21474836480", result.Available)
		}
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
