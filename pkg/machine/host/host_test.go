package host_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// TestFakeHostFS verifies that Fake().FS() reads files from the rootDir.
func TestFakeHostFS(t *testing.T) {
	// Set up a temporary directory with a test file.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "proc"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "proc", "cpuinfo"), []byte("model name\t: test cpu\n"), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	data, err := fs.ReadFile(h.FS(), "proc/cpuinfo")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "model name\t: test cpu\n" {
		t.Errorf("unexpected content: %q", string(data))
	}
}

// TestFakeHostEvalSymlinks verifies that EvalSymlinks follows relative symlinks
// within the fake root and rejects escaping ones.
func TestFakeHostEvalSymlinks(t *testing.T) {
	dir := t.TempDir()

	// Create target: dir/target/real
	if err := os.MkdirAll(filepath.Join(dir, "target"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "target", "real"), []byte("real file"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create relative symlink: dir/link -> target/real
	if err := os.Symlink("target/real", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)

	t.Run("relative symlink resolves", func(t *testing.T) {
		got, err := h.EvalSymlinks("link")
		if err != nil {
			t.Fatalf("EvalSymlinks: %v", err)
		}
		if got != "target/real" {
			t.Errorf("expected target/real, got %q", got)
		}
	})

	t.Run("absolute symlink is rejected", func(t *testing.T) {
		// Create an absolute symlink that escapes the root
		absSymlink := filepath.Join(dir, "escaping")
		if err := os.Symlink("/etc/passwd", absSymlink); err != nil {
			t.Fatal(err)
		}
		_, err := h.EvalSymlinks("escaping")
		if err == nil {
			t.Fatal("expected error for escaping symlink, got nil")
		}
	})
}

// TestFakeHostRunCommandNvidia verifies that RunCommand maps nvidia-smi invocations
// to files under run/nvidia-smi/<slot>/<query>.
func TestFakeHostRunCommandNvidia(t *testing.T) {
	dir := t.TempDir()
	slot := "0000:01:00.0"
	query := "memory.total"

	if err := os.MkdirAll(filepath.Join(dir, "run", "nvidia-smi", slot), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run", "nvidia-smi", slot, query), []byte("4096 MiB"), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	out, err := h.RunCommand(context.Background(), "nvidia-smi", nil,
		"--id="+slot, "--query-gpu="+query, "--format=csv,noheader")
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if string(out) != "4096 MiB" {
		t.Errorf("unexpected output: %q", string(out))
	}
}

// TestFakeHostRunCommandClinfo verifies that RunCommand maps clinfo to run/clinfo.json.
func TestFakeHostRunCommandClinfo(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "run"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run", "clinfo.json"), []byte(`{"devices":[]}`), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	out, err := h.RunCommand(context.Background(), "clinfo", nil, "--json")
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if string(out) != `{"devices":[]}` {
		t.Errorf("unexpected output: %q", string(out))
	}
}

// TestFakeHostRunCommandUnknown verifies that unknown commands return an error.
func TestFakeHostRunCommandUnknown(t *testing.T) {
	h := host.Fake(t.TempDir())
	_, err := h.RunCommand(context.Background(), "unknown-command", nil)
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}

// TestFakeHostStatFs verifies that StatFs reads from run/disk-stats.json.
func TestFakeHostStatFs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "run"), 0755); err != nil {
		t.Fatal(err)
	}
	// JSON keys use leading "/" for human readability.
	json := `{"/var/lib/snapd/snaps": {"total": 53687091200, "avail": 21474836480}}`
	if err := os.WriteFile(filepath.Join(dir, "run", "disk-stats.json"), []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	// API uses io/fs path convention (no leading slash).
	stats, err := h.StatFs("var/lib/snapd/snaps")
	if err != nil {
		t.Fatalf("StatFs: %v", err)
	}
	if stats.Total != 53687091200 {
		t.Errorf("expected total 53687091200, got %d", stats.Total)
	}
	if stats.Avail != 21474836480 {
		t.Errorf("expected avail 21474836480, got %d", stats.Avail)
	}
}

// TestFakeHostStatFsMissingKey verifies that StatFs returns an error for a missing key.
func TestFakeHostStatFsMissingKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "run"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run", "disk-stats.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	h := host.Fake(dir)
	_, err := h.StatFs("var/lib/snapd/snaps")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}
