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

	h := host.Fake(dir)
	results, err := infoWithDirs(h, []string{host.FakeSnapStoragePath, host.FakeHostRoot, host.FakeHostSnap})
	if err != nil {
		t.Fatalf("Info() error: %v", err)
	}

	want := map[string]struct {
		total uint64
		avail uint64
	}{
		host.FakeSnapStoragePath: {total: 219902325555200, avail: 54975581388800},
		host.FakeHostRoot:        {total: 109951162777600, avail: 21990232555520},
		host.FakeHostSnap:        {total: 54975581388800, avail: 10995116277760},
	}

	for _, result := range results {
		if result.MountPoint == nil || *result.MountPoint != "/" {
			t.Errorf("unexpected mount point: got %v, want %q", result.MountPoint, "/")
		}
		exp, ok := want[result.Path]
		if !ok {
			t.Errorf("unexpected path: %q", result.Path)
			continue
		}
		if result.Total != exp.total {
			t.Errorf("Total for %s = %d, want %d", result.Path, result.Total, exp.total)
		}
		if result.Available != exp.avail {
			t.Errorf("Available for %s = %d, want %d", result.Path, result.Available, exp.avail)
		}
	}
}
