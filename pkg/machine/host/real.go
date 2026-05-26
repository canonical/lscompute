package host

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/canonical/lscompute/pkg/machine/types"
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

func (realHost) StatFs(path string) (types.DirStats, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(filepath.Join("/", path), &st); err != nil {
		return types.DirStats{}, fmt.Errorf("statfs %s: %v", path, err)
	}
	return types.DirStats{
		Total: st.Blocks * uint64(st.Bsize),
		Avail: st.Bavail * uint64(st.Bsize),
	}, nil
}
