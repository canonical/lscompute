package host

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type realHost struct{}

func (realHost) FS() fs.FS { return os.DirFS("/") }

func (realHost) EvalSymlinks(path string) (string, error) {
	abs, err := filepath.EvalSymlinks(filepath.Join("/", path))
	if err != nil {
		return "", err
	}
	rel := strings.TrimPrefix(abs, "/")
	if rel == "" {
		rel = "." // io/fs convention for root
	}
	return rel, nil
}

func (realHost) RunCommand(ctx context.Context, name string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	return cmd.Output()
}

func (realHost) StatFs(path string) (dirStats, error) {
	var st unix.Statfs_t
	fullPath := filepath.Join("/", path)
	if err := unix.Statfs(fullPath, &st); err != nil {
		return dirStats{}, fmt.Errorf("statfs %s: %w", path, err)
	}

	mountpoint, err := getMountpoint(fullPath)
	if err != nil {
		// Fall back to the path itself if we can't determine the mountpoint
		mountpoint = fullPath
	}

	return dirStats{
		Mountpoint: mountpoint,
		Total:      st.Blocks * uint64(st.Bsize),
		Avail:      st.Bavail * uint64(st.Bsize),
	}, nil
}

// getMountpoint retrieves the actual mountpoint for a given path by parsing /proc/mounts.
func getMountpoint(path string) (string, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var longestMatch string
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		mountpoint := parts[1]
		// Find the longest matching mountpoint (most specific). Ensure we match a full path segment.
		if (mountpoint == "/" || path == mountpoint || strings.HasPrefix(path, mountpoint+"/")) && len(mountpoint) > len(longestMatch) {
			longestMatch = mountpoint
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if longestMatch == "" {
		return "", fmt.Errorf("mountpoint not found for %s", path)
	}
	return longestMatch, nil
}
