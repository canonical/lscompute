// Package host defines the seam between production and test code.
// Real() returns an implementation rooted at the live filesystem;
// Fake() returns one rooted at a per-machine fake rootfs that reads
// pre-captured command output and statfs values from files.
package host

import (
	"context"
	"io/fs"
)

// Host is the seam between production and tests. Real() returns an implementation
// rooted at the live filesystem; Fake() returns one rooted at a per-machine fake
// rootfs that reads pre-captured command output and statfs values from files.
//
// Path conventions for FS() and EvalSymlinks match io/fs:
//   - Forward slashes, no leading "/" (e.g. "sys/bus/pci/devices", not "/sys/...").
//   - Use "." for the root itself.
//
// StatFs follows the same io/fs path convention as FS() and EvalSymlinks at the
// API boundary: no leading slash (e.g. "var/lib/snapd/snaps"). Both Real() and
// Fake() prepend "/" internally — Real() to build an absolute path for
// unix.Statfs(2), Fake() to look up the matching key in its JSON fixture (which
// is hand-authored with leading-slash keys for human readability).
//
// Symlinks in a fake rootfs MUST be relative. An absolute symlink would escape the
// fake root and reach the real filesystem; EvalSymlinks rejects targets that
// resolve outside Root.
type Host interface {
	// FS returns a read-only filesystem view of the host. Callers use the
	// standard io/fs helpers on it: fs.ReadFile, fs.ReadDir, fs.WalkDir, etc.
	FS() fs.FS

	// EvalSymlinks resolves a symlink and returns the target as a path relative
	// to the host's root (same path convention as FS()).
	EvalSymlinks(path string) (string, error)

	// RunCommand executes an external command and returns its stdout. Real()
	// shells out via os/exec with a context-bound timeout, kill-tree cancel hook,
	// and `env` appended to os.Environ() (entries are "KEY=VALUE"; pass nil to
	// inherit unchanged). Fake() ignores ctx and env and maps the invocation to
	// a pre-recorded file under <root>/run/.
	RunCommand(ctx context.Context, name string, env []string, args ...string) ([]byte, error)

	// StatFs returns total and available bytes for a directory. Path follows
	// io/fs convention at the API: no leading slash (e.g. "var/lib/snapd/snaps").
	// Real() prepends "/" before calling unix.Statfs(2). Fake() reads canned
	// values from <root>/run/disk-stats.json, whose keys *do* have a leading "/"
	// (so the fixture reads like absolute paths on a real host); Fake() prepends
	// "/" to the API path before looking up.
	StatFs(path string) (DirStats, error)
}

// Real returns a Host that talks to the live system.
func Real() Host { return &realHost{} }

// Fake returns a Host rooted at the given directory, which should be the
// machine-root directory (e.g. "test_data/machines/xps13-7390/machine-root").
// RunCommand maps invocations to pre-captured output under <rootDir>/run/<name>/...;
// StatFs reads <rootDir>/run/disk-stats.json.
func Fake(rootDir string) Host { return &fakeHost{root: rootDir} }
