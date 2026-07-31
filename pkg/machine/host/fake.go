package host

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	FakeHostRoot        = "/fakehost/root"
	FakeHostSnap        = "/fakehost/snap"
	FakeSnapStoragePath = "/fakehost/var/lib/snapd/snaps"
)

type fakeHost struct{ root string }

func (h *fakeHost) FS() fs.FS { return os.DirFS(h.root) }

func (h *fakeHost) EvalSymlinks(path string) (string, error) {
	absRoot, err := filepath.Abs(h.root)
	if err != nil {
		return "", err
	}
	abs, err := filepath.EvalSymlinks(filepath.Join(absRoot, path))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, abs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("symlink %q escapes fake root", path)
	}
	return rel, nil
}

// RunCommand maps a command invocation to a pre-recorded file under <root>/run/.
//
// Mapping rules:
//
//	nvidia-smi --id=<slot> --query-gpu=<query> ... → run/nvidia-smi/<slot>/<query>
//	clinfo --json                                   → run/clinfo.json
//
// ctx and env are ignored in tests. Commands without a mapping return an error.
func (h *fakeHost) RunCommand(_ context.Context, name string, _ []string, args ...string) ([]byte, error) {
	var filePath string

	switch name {
	case "nvidia-smi":
		slot, query, err := parseNvidiaSmiArgs(args)
		if err != nil {
			return nil, fmt.Errorf("fake RunCommand: nvidia-smi: %w", err)
		}
		filePath = filepath.Join(h.root, "run", "nvidia-smi", slot, query)

	case "clinfo":
		// All clinfo invocations map to the same captured JSON output.
		filePath = filepath.Join(h.root, "run", "clinfo.json")

	default:
		return nil, fmt.Errorf("fake RunCommand: no mapping for command %q", name)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("fake RunCommand: reading %s: %w", filePath, err)
	}
	return data, nil
}

// parseNvidiaSmiArgs extracts the PCI slot (from --id=<slot>) and the query name
// (from --query-gpu=<query>) from the nvidia-smi argument list.
func parseNvidiaSmiArgs(args []string) (slot, query string, err error) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--id=") {
			slot = strings.TrimPrefix(arg, "--id=")
		} else if strings.HasPrefix(arg, "--query-gpu=") {
			query = strings.TrimPrefix(arg, "--query-gpu=")
		}
	}
	if slot == "" {
		return "", "", fmt.Errorf("missing --id= flag")
	}
	if query == "" {
		return "", "", fmt.Errorf("missing --query-gpu= flag")
	}
	return slot, query, nil
}

func (h *fakeHost) DirStat(path string) (*unix.Statfs_t, error) {
	var st unix.Statfs_t
	st.Bsize = int64(1024) // block size in bytes
	switch path {
	case FakeHostRoot:
		st.Blocks = 100 * 1024 * 1024 * uint64(st.Bsize) // 100 GiB
		st.Bavail = 20 * 1024 * 1024 * uint64(st.Bsize)  // 20 GiB
	case FakeHostSnap:
		st.Blocks = 50 * 1024 * 1024 * uint64(st.Bsize) // 50 GiB
		st.Bavail = 10 * 1024 * 1024 * uint64(st.Bsize) // 10 GiB
	case FakeSnapStoragePath:
		st.Blocks = 200 * 1024 * 1024 * uint64(st.Bsize) // 200 GiB
		st.Bavail = 50 * 1024 * 1024 * uint64(st.Bsize)  // 50 GiB
	default:
		return nil, fmt.Errorf("fake DirStat: no mapping for path %q", path)
	}

	return &st, nil
}

func (h *fakeHost) GetMountPoint(path string) (string, error) {
	return "/", nil
}

func (h *fakeHost) GetDirectories() []string {
	return []string{FakeSnapStoragePath}
}
