package host

import (
	"context"
	"io/fs"

	"golang.org/x/sys/unix"
)

type Host interface {
	FS() fs.FS

	EvalSymlinks(path string) (string, error)

	RunCommand(ctx context.Context, name string, env []string, args ...string) ([]byte, error)

	DirStat(path string) (*unix.Statfs_t, error)

	GetMountPoint(path string) (string, error)

	GetDirectories() []string
}

func Real() Host { return &realHost{} }

func Fake(rootDir string) Host { return &fakeHost{root: rootDir} }
