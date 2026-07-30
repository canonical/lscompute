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

const (
	snapStoragePath = "/var/lib/snapd/snaps"
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

func (realHost) DirStat(path string) (*unix.Statfs_t, error) {
	var st unix.Statfs_t
	fullPath := filepath.Join("/", path)
	if err := unix.Statfs(fullPath, &st); err != nil {
		return nil, fmt.Errorf("statfs %s: %w", path, err)
	}
	return &st, nil
}

// GetMountpoint retrieves the actual mountpoint for a given path by parsing the
// host's mount table.
//
// It reads /proc/1/mounts (PID 1, the host init) rather than /proc/mounts
// (a.k.a. /proc/self/mounts). When lscompute runs inside a strict snap, the
// process has its own mount namespace in which the host root "/" is not present
// and /var/lib/snapd/snaps appears as a bind mount; parsing /proc/self/mounts
// there would report /var/lib/snapd/snaps instead of the real host mountpoint.
// PID 1 lives in the host's root mount namespace, so its mount table reflects
// the real filesystem layout. On a non-snap host /proc/1/mounts is equivalent
// to /proc/self/mounts. Reading another PID's mount table requires the
// mount-observe interface under confinement.
func (realHost) GetMountpoint(path string) (string, error) {
	file, err := os.Open("/proc/1/mounts")
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

func (realHost) GetDirectories() []string {
	return []string{snapStoragePath}
}
