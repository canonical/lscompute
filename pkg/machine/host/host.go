package host

import (
	"context"
	"io/fs"
)

type DirStat struct {
	Total     uint64
	Available uint64
}

type Host interface {
	FS() fs.FS

	EvalSymlinks(path string) (string, error)

	RunCommand(ctx context.Context, name string, env []string, args ...string) ([]byte, error)

	DirStat(path string) (*DirStat, error)

	GetMountPoint(path string) (string, error)

	GetDirectories() []string
}

func Real() Host { return &realHost{} }

func Fake(rootDir string) Host { return &fakeHost{root: rootDir} }
