package host

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

// StatFs reads disk stats from <root>/run/disk-stats.json.
// The JSON keys carry a leading "/" (human-readable absolute paths);
// fakeHost prepends "/" to the API path before looking up, matching
// what realHost does when building the unix.Statfs argument.
func (h *fakeHost) StatFs(path string) (dirStats, error) {
	filePath := filepath.Join(h.root, "run", "disk-stats.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return dirStats{}, fmt.Errorf("fake StatFs: reading %s: %w", filePath, err)
	}

	var stats map[string]dirStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return dirStats{}, fmt.Errorf("fake StatFs: parsing %s: %w", filePath, err)
	}

	// JSON keys have leading "/" for readability; prepend "/" to match.
	key := "/" + path
	entry, ok := stats[key]
	if !ok {
		return dirStats{}, fmt.Errorf("fake StatFs: no entry for %q in %s", key, filePath)
	}
	return entry, nil
}
