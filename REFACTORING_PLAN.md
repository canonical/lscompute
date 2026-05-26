# Refactoring Plan: Per-Machine Fake Rootfs + Single `host.Host` Seam

## Goal

Make the codebase testable end-to-end against real-world data captured from real machines,
using a single idiomatic Go seam. One code path runs in both production and tests.

The user-stated vision drives the plan: **one directory per machine under
`test_data/machines/`, containing a fake rootfs and any captured command output. Tests drive
the full pipeline against that directory; production runs against `/` on the real host.**

---

## Why the previous plan was wrong

The previous `SystemCaller` interface enumerated every kind of system call (`ProcCpuInfo`,
`LsPci`, `NvidiaSmi`, `ReadSysFile`, …) as separate methods. The problem wasn't
"interface" — it was the *granularity*:

- **Catalogue interface, not a small one.** Idiomatic Go interfaces are small and describe
  a *capability* (read, exec, …), not an enumeration of every system call by name.
- **Brittle.** Every new data source adds a method on the interface.
- **Redundant.** Most of the code already reads regular files; only two paths actually run
  external commands (`nvidia-smi`, `clinfo`).

Recent commits (`Remove dependency on lspci and lsusb`, `Remove dependency on pciid db lib`)
have already made the codebase almost entirely file-driven. The new design embraces that:
a small four-method interface where each method is a distinct kind of capability
(filesystem, symlink resolution, command exec, statfs syscall), built on `io/fs.FS` for the
filesystem half.

---

## The new seam: `host.Host`

A small interface with four methods, passed into every package entry point:

```go
// pkg/machine/host/host.go
package host

import (
    "context"
    "io/fs"

    "github.com/canonical/lscompute/pkg/machine/types"
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
    StatFs(path string) (types.DirStats, error)
}

// Real returns a Host that talks to the live system.
func Real() Host { return &realHost{} }

// Fake returns a Host rooted at the given directory, which should be the
// machine-root directory (e.g. "test_data/machines/xps13-7390/machine-root").
// RunCommand maps invocations to pre-captured output under <rootDir>/run/<name>/...;
// StatFs reads <rootDir>/run/disk-stats.json.
// See the test data layout and command output convention sections below.
func Fake(rootDir string) Host { return &fakeHost{root: rootDir} }
```

```go
// pkg/machine/host/real.go
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
```

```go
// pkg/machine/host/fake.go (sketch — see test data layout for file conventions)
package host

import (
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

// RunCommand and StatFs implementations are described in the convention sections.
```

Callers do file reads via the standard `io/fs` helpers, e.g.:

```go
data, err := fs.ReadFile(h.FS(), "proc/cpuinfo")
entries, err := fs.ReadDir(h.FS(), "sys/bus/pci/devices")
```

Three properties make this work:

1. **Small interface; concrete implementations are private.** Four methods, each a
   distinct capability (filesystem, symlink resolution, command exec, statfs syscall).
   `realHost` and `fakeHost` are package-private — callers see only the `Host`
   interface, which makes substitution and mocking trivial.
2. **Built on `io/fs.FS`.** File and directory reads go through the standard library's
   filesystem abstraction. Callers use `fs.ReadFile(h.FS(), path)`, `fs.ReadDir(...)`,
   `fs.WalkDir(...)`, and any other `io/fs` consumer. Path arguments follow the `io/fs`
   convention (no leading slash, "." for root). No circular imports — every package
   accepts `host.Host` by interface value.
3. **Symlinks work transparently — when they're relative.** `EvalSymlinks` operates
   under the host's root, so existing test data (which uses *relative* symlinks in
   `machine-root/sys/class/drm/…`) keeps working. The `fakeHost.EvalSymlinks`
   implementation actively rejects targets that escape Root.

   Caveat: `os.DirFS` (which backs `FS()`) is **not** a sandbox — a stray absolute or
   escaping symlink in a fake rootfs would be silently followed by `fs.ReadFile` /
   `fs.ReadDir`. We rely on convention (committed test data uses only relative
   symlinks) for `FS()` reads; `EvalSymlinks` is the only path that actively rejects
   escapes. This is fine because we control the test data, but it means the
   "no absolute symlinks" rule is enforced by code review for `FS()` callers.

---

## Test data layout (one fake rootfs per machine)

Everything for a machine lives under `test_data/machines/<machine>/machine-root/`.
The arrows below show what each file mirrors on a real Linux host; Go callers reference
these same paths without a leading slash (e.g. `fs.ReadFile(h.FS(), "proc/cpuinfo")`).

```
test_data/machines/<machine>/
    machine-root/
        proc/cpuinfo                       ← /proc/cpuinfo
        proc/meminfo                       ← /proc/meminfo
        proc/sys/kernel/arch               ← single-line machine arch (Linux 6.1+ sysctl)
        sys/bus/pci/devices/<slot>/        ← vendor, device, class, subsystem_*
        sys/bus/usb/devices/<bus>-<port>/  ← idVendor, idProduct, busnum, devnum (USB machines only)
        sys/class/kfd/kfd/topology/nodes/  ← AMD only
        sys/class/drm/renderD<n>           ← relative symlink, AMD only
        usr/share/misc/pci.ids             ← optional; if present, friendly-name tests use it
        usr/share/misc/usb.ids             ← optional
        run/disk-stats.json                ← JSON map of path → {total, avail} (statfs fixture)
        run/nvidia-smi/<slot>/<query>      ← per-query nvidia-smi capture (NVIDIA only)
        run/clinfo.json                    ← captured `clinfo --json` (Intel GPU only)

    hardware-info.json                     ← golden file (only for curated machines)
```

The `run/` directory holds both **captured command output** (`nvidia-smi`, `clinfo`) and
**hand-authored JSON fixtures** (`disk-stats.json`) — both feed the non-FS seams
(`RunCommand`, `StatFs`). It's the inverse of `machine-root/sys`, `machine-root/proc`,
etc., which mirror real kernel paths verbatim.

The flat-text helpers (`cpuinfo.txt`, `meminfo.txt`, `uname-m.txt`, `disk.txt`, `lspci.txt`,
`additional-properties.json`) currently sitting at the machine root are migrated **into**
`machine-root/` and the originals deleted. Mapping:

- `cpuinfo.txt` → `machine-root/proc/cpuinfo`
- `meminfo.txt` → `machine-root/proc/meminfo`
- `uname-m.txt` → `machine-root/proc/sys/kernel/arch`
- `disk.txt` → `machine-root/run/disk-stats.json` (hand-authored from the captured df numbers)
- `lspci.txt` → expanded into a sysfs mirror under `machine-root/sys/bus/pci/devices/<slot>/` by a Python one-shot at `test_data/tools/lspci_to_sysfs.py`

The Python script stays in the repo so we can regenerate if needed and so it documents the
mapping. It is intentionally **not** Go: keeping it out of the main Go module avoids polluting
`go build ./...` / `go test ./...`, and the lspci-line-to-sysfs-attribute mapping reads more
naturally as text munging in Python than in Go.

The `additional-properties.json` files are also deleted; the golden `hardware-info.json` files
become the single source of truth for expected output.

### Command output convention

`Fake().RunCommand(ctx, name, env, args...)` maps a command invocation to a file under
`<root>/run/`. `ctx` and `env` are ignored in tests. The mapping is intentionally simple:

| Command | File |
|---|---|
| `nvidia-smi --id=0000:01:00.0 --query-gpu=memory.total --format=…` | `run/nvidia-smi/0000:01:00.0/memory.total` |
| `nvidia-smi --id=0000:01:00.0 --query-gpu=compute_cap --format=…` | `run/nvidia-smi/0000:01:00.0/compute_cap` |
| `clinfo --json` | `run/clinfo.json` |

`Fake.RunCommand` parses the `--id=` and `--query-gpu=` flags from `args` to find the right file
for `nvidia-smi`. Other commands map to a static path keyed off `name`. The mapping is a tiny
switch statement in `host/fake.go`; commands without a mapping return a "not found" error so
mis-recorded test data is caught immediately rather than silently producing empty output.

`disk.Info()` does not go through `RunCommand` — it uses `Host.StatFs` instead, which in `Fake`
reads `run/disk-stats.json`:

```json
{
    "/var/lib/snapd/snaps": {"total": 53687091200, "avail": 21474836480}
}
```

JSON keys carry a leading `/` so the fixture reads like absolute paths on a real
host. `fakeHost.StatFs` prepends `/` to its API argument before looking up the key,
which mirrors what `realHost.StatFs` does to build the argument for `unix.Statfs(2)`.

---

## Function signature changes

Every package entry point takes a `host.Host` (interface) value. Internal helpers that
already accept a `rootDir string` (the AMD package) are rewritten to take `host.Host`.
The `Info()` / `Devices()` shape stays the same.

| Package | Before | After |
|---|---|---|
| `cpu` | `Info() ([]CpuInfo, error)` | `Info(h host.Host) ([]CpuInfo, error)` |
| `memory` | `Info() (MemoryInfo, error)` | `Info(h host.Host) (MemoryInfo, error)` |
| `disk` | `Info() (map[string]DirStats, error)` | `Info(h host.Host) (map[string]DirStats, error)` |
| `pci` | `Devices(includeFriendly bool)` | `Devices(h host.Host, includeFriendly bool)` |
| `usb` | `Devices(includeFriendly bool)` | `Devices(h host.Host, includeFriendly bool)` |
| `pci/amd` | `AdditionalProperties(device)` | `AdditionalProperties(h host.Host, device)` |
| `pci/nvidia` | `AdditionalProperties(device)` | `AdditionalProperties(h host.Host, device)` |
| `pci/intel` | `AdditionalProperties(device)` | `AdditionalProperties(h host.Host, device)` |
| `machine` (`devices.go`) | `Devices(friendly bool)` | `Devices(h host.Host, friendly bool)` |
| `machine` (`machine.go`) | `Get(friendly bool)` | `Get(h host.Host, friendly bool)` |

`host.Host` is consistently the **first** parameter on every entry point, matching the
`context.Context` convention. `cmd/lscompute/main.go` becomes a one-line change:
`machine.Get(host.Real(), true)`.

### What gets deleted

- `cpu.hostProcCpuInfo`, `cpu.hostMachineArch`, `cpu.InfoFromRawData`
- `memory.hostProcMemInfo`, `memory.InfoFromRawData`
- `disk.hostDf`, `disk.parseDf`, `disk.statFs`, `disk.InfoFromRawData`, `disk_df.go`, `syscall_statfs.go` — `Info(h)` calls `h.StatFs(dir)` directly; the `statfs(2)` logic moves into `host.Real().StatFs`
- `pci.hostSysPci`
- `usb.hostSysUsb`
- `amd.gpuProperties` / `amd.gpuPropertiesFromDir` collapse into one function taking `Host`
- `nvidia.nvidiaSmi` (the private exec wrapper) — replaced by `h.RunCommand(ctx, "nvidia-smi", []string{"LANG=C"}, args...)`
- `intel.vRam`'s `exec.Command("clinfo")` block — replaced by `h.RunCommand(ctx, "clinfo", nil, "--json")`

### What stays untouched

All remaining pure-parse functions: `parseProcCpuInfo*`, `parseProcMemInfo`, `parseClinfoJson`,
`parseVramAmount`, `parseGfxTargetVersion`, `splitPciIdName`, `splitSubsystemLine`, etc. They
are already isolated and already covered by `_test.go` files. They get one more user (`host.Host`-based
callers) but their signatures don't change. (`parseDf` is the exception — it goes away with the
df code path.)

### pci.ids / usb.ids lookup

`findPciIdsFile()` and `findUsbIdsFile()` walk a hard-coded list of system paths. They are
extended to take a `host.Host` and probe candidates with `fs.Stat(h.FS(), path)`.
`lookupPciIds()` is changed from streaming via `bufio.Scanner(os.Open(...))` to reading the
whole file with `fs.ReadFile(h.FS(), path)` and scanning the byte slice — `pci.ids` is a few
MB at most, and routing through the `io/fs` seam is worth the in-memory copy.

For the **three machines with golden files** (`raspberry-pi-5`, `raspberry-pi-5+hailo-8`,
`xps13-7390`), we ship a **curated mini `pci.ids`** under
`machine-root/usr/share/misc/pci.ids` containing only the vendor/device entries those machines
actually need. Same idea for `usb.ids` if the machine has USB devices. The curated files are
authored by hand (extract the relevant vendor blocks from the system `pci.ids`) — at most a
handful of devices per machine, so this stays trivial. This keeps the test data self-contained:
the golden files include friendly names, and the test result does not depend on which
`pci.ids` happens to be installed on the developer's machine.

Other machines (no golden file, parse-only test) don't need a `pci.ids`; the friendly-name
lookup produces warnings (logged via `t.Log`, not asserted on) and the device's friendly-name
fields stay unset. The full-pipeline test runs with `friendlyNames=true` to exercise the
lookup code path.

---

## Test strategy

Three layers, each cheap to maintain:

### 1. Pure-parse unit tests (already present, mostly keep as-is)

Tests like `parseGfxTargetVersion`, `parseClinfoJson`, `parseProcCpuInfoArm64`
take strings in and assert structured output. These are the fastest and most stable tests.
Keep them. They don't need a `host.Host` at all. (The `parseDf` test goes away with the df
code path.)

### 2. Per-package tests using `host.Fake(...)`

Each package gets a small test that drives its entry point against the curated machine
directories. Pattern:

```go
func TestCpuInfo(t *testing.T) {
    cases := []struct {
        machine  string
        wantArch string
    }{
        {"raspberry-pi-5", "arm64"},
        {"xps13-7390", "amd64"},
        {"hp-zbook-i712850HX+RadeonPROW6600M", "amd64"},
    }
    for _, tc := range cases {
        t.Run(tc.machine, func(t *testing.T) {
            h := host.Fake("../../../test_data/machines/" + tc.machine + "/machine-root")
            cpus, err := Info(h)
            // assert…
        })
    }
}
```

Each package picks the 2-3 machines that meaningfully exercise its code (e.g. AMD tests
pick the two AMD-GPU machines; NVIDIA picks the GTX-1080Ti machine; Intel picks one of the
Arc machines).

### 3. Full-pipeline golden test in the `machine` package

One table-driven test iterates every directory in `test_data/machines/` and runs the entire
pipeline:

```go
var update = flag.Bool("update", false, "rewrite golden files instead of asserting")

func TestGetFromMachineDirs(t *testing.T) {
    entries, _ := os.ReadDir("../../test_data/machines")
    for _, entry := range entries {
        if !entry.IsDir() { continue }
        t.Run(entry.Name(), func(t *testing.T) {
            dir := "../../test_data/machines/" + entry.Name()
            h := host.Fake(filepath.Join(dir, "machine-root"))

            got, warnings, err := Get(h, false)
            if err != nil { t.Fatal(err) }
            for _, w := range warnings { t.Log("warning:", w) }

            goldenPath := filepath.Join(dir, "hardware-info.json")
            if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
                return // parsing-only check for machines without a golden
            }

            if *update {
                writeGolden(t, goldenPath, got)
                return
            }
            assertEqualToGolden(t, goldenPath, got)
        })
    }
}
```

- **Every machine** under `test_data/machines/` is exercised end-to-end. If any machine's
  raw data trips a parser, the test fails — even without a golden file.
- **Golden-file comparison** only runs for machines that have one. Currently three: `raspberry-pi-5`,
  `raspberry-pi-5+hailo-8`, `xps13-7390`. New goldens are added by running `go test -update`
  for a specific machine after manually verifying the output is correct.
- **Warnings are logged, not asserted on.** A machine missing `pci.ids` in its fake rootfs
  produces a friendly-name warning; that's fine, friendly names are off in golden tests.

This is the central test. It is the same code path as `cmd/lscompute/main.go` — the only
difference is `host.Fake(dir)` versus `host.Real()`. There is no second pipeline to maintain.

### Coverage targets

After this refactor, `go test ./... -cover` should show meaningful coverage in every
sub-package. The realistic targets are:

| Package | Target | What's left uncovered |
|---|---|---|
| `cpu` | 80%+ | error branches on malformed `/proc/cpuinfo` |
| `memory` | 90%+ | error branches |
| `disk` | 90%+ | error branches; `statfs(2)` itself runs in `host.Real()` |
| `pci`, `usb` | 80%+ | symlink-walk error branches |
| `pci/amd` | 70%+ | sysfs error branches (no exec; AMD reads `/sys/bus/pci` + `/sys/class/kfd`) |
| `pci/nvidia`, `pci/intel` | 70%+ | command-error branches in `Real()` |
| `machine` | 90%+ | covered transitively by the full-pipeline test |
| `host` | 90%+ | `Real()`'s exec and statfs branches are the only gaps |

We don't chase coverage in `host.Real()`'s exec branch — exercising it would require a real
`nvidia-smi`/`clinfo` on the test runner, which we deliberately don't have. `Real().StatFs` is
trivially exercisable on any Linux box; leaving it uncovered is a question of taste, not
feasibility.

---

## Migration order

Each step compiles and tests pass before the next step starts. No "everything is broken
for two weeks" intermediate states.

1. **Add `pkg/machine/host/host.go` + `host/real.go` + `host/fake.go`.** Self-contained,
   no other code uses it yet. Add `host/host_test.go` covering `FS()` reads (via
   `fs.ReadFile`/`fs.ReadDir`), `EvalSymlinks` (including the "rejects escaping symlinks"
   case for `fakeHost`), and `Fake.RunCommand`'s command-to-file mapping.
2. **Migrate `cpu` to take `host.Host`.** Update `machine.Get()` to pass it through. Migrate
   the existing test data: `cpuinfo.txt` → `machine-root/proc/cpuinfo`, `uname-m.txt` →
   `machine-root/proc/sys/kernel/arch`. Repoint the existing `proc_cpuinfo_test.go` glob from
   `cpuinfo.txt` to `machine-root/proc/cpuinfo`, or fold it into the new `host.Fake`-driven
   `cpu_test.go` — pick one; don't keep two tests pointing at different file conventions.
3. **Migrate `memory` the same way.** `meminfo.txt` → `machine-root/proc/meminfo`. Repoint
   `proc_meminfo_test.go` the same way.
4. **Migrate `disk`.** Single `Info(h)` path: call `h.StatFs(dir)` for each watched directory.
   Convert the existing `disk.txt` (captured df output) into `machine-root/run/disk-stats.json` —
   a hand-authored JSON map of directory path → `{"total": <bytes>, "avail": <bytes>}`. Delete
   `disk_df.go`, `syscall_statfs.go`, `disk.InfoFromRawData`, and `parseDf`. The `statfs(2)`
   syscall stays — it just moves into `host.Real().StatFs` so the fake-rootfs path can
   substitute a JSON-backed implementation.
5. **Migrate `pci` core** (the sysfs read in `syspci.go`). `os.ReadDir("/sys/bus/pci/devices")`
   becomes `fs.ReadDir(h.FS(), "sys/bus/pci/devices")` (no leading slash). Per-attribute
   reads become `fs.ReadFile(h.FS(), filepath.Join("sys/bus/pci/devices", slot, "vendor"))`.
   Write a one-shot **Python** script at
   `test_data/tools/lspci_to_sysfs.py` that takes a machine directory and converts its
   `lspci.txt` into a sysfs mirror under
   `machine-root/sys/bus/pci/devices/<slot>/{vendor,device,class,subsystem_vendor,subsystem_device}`.
   Run it once across every machine, commit the generated sysfs trees, delete `lspci.txt`. The
   script is Python (not Go) so it stays out of `go build ./...` / `go test ./...` and is
   unambiguously not part of the runtime — it's documentation-as-code for the mapping.
   Note: `lspci.txt` formats vary across captures (older outputs may omit the `0000:` PCI
   domain prefix on slot IDs; some include programming-interface bytes, some don't). The
   script must normalize to the kernel's canonical `0000:bb:dd.f` slot form. Spot-check the
   generated sysfs tree for two or three machines after the bulk run before deleting the
   originals.
6. **Migrate `usb`** the same way.
7. **Migrate `pci/amd`.** The existing `rootDir string` parameter is replaced by `host.Host`.
   Existing AMD tests stay structurally the same — they just construct `host.Fake(...)` instead
   of passing strings. The two machines with AMD `machine-root/` data already work.
8. **Migrate `pci/nvidia`.** Replace `nvidiaSmi(args...)` with
   `h.RunCommand(ctx, "nvidia-smi", []string{"LANG=C"}, args...)` (callers own the
   `context.WithTimeout` they pass). Capture real `nvidia-smi` output from one NVIDIA
   machine into `machine-root/run/nvidia-smi/<slot>/<query>` files (e.g. on the
   `i5-3570k+arc-a580+gtx1080ti` machine).
9. **Migrate `pci/intel`.** Replace `exec.Command("clinfo", "--json")` with
   `h.RunCommand(ctx, "clinfo", nil, "--json")`. Drop one of the `test_data/clinfo/*.json`
   fixtures into the right machine's `machine-root/run/clinfo.json`.
10. **Migrate `machine.Get()` + `cmd/lscompute/main.go`.** Add `host.Host` parameter; delete
    `GetFromRawData`/`InfoFromRawData` shells if any remain.
11. **Write the full-pipeline `TestGetFromMachineDirs`** in `pkg/machine`. Add the `-update`
    flag for golden regeneration.
12. **Add curated `pci.ids` for golden machines.** For each of `raspberry-pi-5`,
    `raspberry-pi-5+hailo-8`, `xps13-7390`, extract **by hand** just the vendor/device entries
    the machine references from the system `pci.ids` and write to
    `machine-root/usr/share/misc/pci.ids`. Same for `usb.ids` if the machine has USB devices.
    Authoring by hand is fine — there are at most a handful of devices per machine. Then
    regenerate the goldens with `go test -update` and verify the friendly names look right.
13. **Cleanup pass.** Delete the now-unused flat files at the machine root level: `lspci.txt`,
    `cpuinfo.txt`, `meminfo.txt`, `uname-m.txt`, `disk.txt`, and the `additional-properties.json`
    files (no longer needed — the sysfs/run mirrors are the source of truth, golden files
    document expected output). Remove any dead helpers left over.

Steps 1-3 are quick (each is ~1 hour of work). Step 4's code change is also ~1 hour, but
the data-entry of converting `disk.txt` → `disk-stats.json` across ~23 machines is grindy;
budget a separate session for it (or convert only the few golden machines first and let
the rest be parse-only until later). Steps 5-6 take more effort because of the sysfs-mirror
generation. Steps 7-9 are mechanical given the data already exists for AMD and exists
separately under `test_data/clinfo/`.

---

## Decisions

- **Step 5 test data conversion**: a Python one-shot at `test_data/tools/lspci_to_sysfs.py`
  converts `lspci.txt` into the sysfs mirror per machine. Stays in the repo for regeneration
  and as documentation of the mapping. Kept out of Go so it doesn't pollute `go build ./...`.
- **Disk info**: production uses `statfs(2)` via `host.Real().StatFs`; tests read from
  `machine-root/run/disk-stats.json` (path → `{total, avail}`). The unused `disk.hostDf` /
  `parseDf` code path is removed. `df` brought process spawn, locale handling, parsing, and
  a snap-confinement requirement for the super-privileged `mount-observe` interface — none of
  which are worth keeping when the syscall already works in prod and the JSON fixture is
  trivial to author.
- **`additional-properties.json` files**: deleted in the cleanup pass. Golden `hardware-info.json`
  files are the single source of truth for expected output.
- **Friendly names in goldens**: yes. Golden machines get a curated mini `pci.ids` (and `usb.ids`
  where relevant) under `machine-root/usr/share/misc/` — authored by hand. Full-pipeline test
  runs with `friendlyNames=true`. Non-golden machines log warnings for missing names; that's fine.
- **`pci.ids` parsing**: switched from streaming `bufio.Scanner(os.Open(...))` to whole-file
  `fs.ReadFile(h.FS(), path)` so the read goes through the `io/fs` seam. The file is a few MB
  at most so the memory cost is negligible.
- **`Host.RunCommand` signature**: takes `(ctx context.Context, name string, env []string,
  args ...string)`. `ctx` preserves the timeouts/kill-tree behavior currently on `nvidia-smi`
  and `clinfo`; `env` preserves the `LANG=C` / `LC_ALL=POSIX` overrides. Both are ignored by
  `Fake`. Going with `(ctx, name, env, args...)` rather than an options struct keeps the
  signature ctx-first like the rest of the standard library.
- **Interface, not struct-with-func-fields.** `Host` is a four-method Go interface; `realHost`
  and `fakeHost` are unexported concrete implementations. This is the conventional Go shape
  for a swappable seam — easy to mock, easy to extend with a third implementation, and the
  call sites read naturally (`h.FS()`, `h.RunCommand(...)`). The earlier struct-with-func-fields
  design was rejected as un-idiomatic.
- **Filesystem reads go through `io/fs.FS`.** `Host` exposes the filesystem half of itself via
  `FS() fs.FS`, and callers use the standard `fs.ReadFile` / `fs.ReadDir` / `fs.WalkDir`
  helpers. `realHost.FS()` returns `os.DirFS("/")`; `fakeHost.FS()` returns `os.DirFS(root)`
  where root is the machine-root directory passed to `Fake()`. Reusing the stdlib abstraction
  means we get every existing `io/fs` consumer for free and we don't reinvent path semantics.
- **Go API paths follow `io/fs` convention throughout** (no leading slash, "." for root) —
  including `StatFs`. This gives the interface a single, consistent path convention.
  `Real().StatFs` prepends "/" internally before calling `unix.Statfs(2)`; callers never see
  the leading slash. This deviates from the earlier "always start with /" rule — the `io/fs`
  convention is the stdlib standard and trying to use leading slashes with `os.DirFS` produces
  invalid-path errors.
- **Hand-authored JSON fixtures use leading-slash keys.** `disk-stats.json` keys look like
  absolute paths (`/var/lib/snapd/snaps`) because the file is read by humans, not by
  `os.DirFS`. `fakeHost.StatFs` prepends "/" to its API argument before lookup, the same way
  `realHost.StatFs` prepends "/" before the syscall. The io/fs convention applies to the Go
  API surface, not to test data that represents real-host absolute paths.
- **`Fake()` takes the machine-root directory directly** (e.g.
  `test_data/machines/xps13-7390/machine-root`). Callers append `/machine-root`; the golden
  test computes `goldenPath` from the parent `dir` before constructing the `Host`. This keeps
  `fakeHost` simple: `FS()` is just `os.DirFS(h.root)` and RunCommand / StatFs resolve paths
  relative to the same root without any extra subdirectory mapping.

---

## Summary

- One small interface (`host.Host`): `FS()`, `EvalSymlinks`, `RunCommand`, `StatFs`. Built
  on `io/fs.FS` for the filesystem half; concrete `realHost` / `fakeHost` types are private.
- One code path: same `machine.Get(h, friendly)` runs in production and tests.
- One test data convention: everything under `machine-root/`, mirroring real `/`.
- One central golden test that exercises every captured machine.

This embraces the direction the codebase is already going (sysfs over external commands)
and uses the conventional Go shapes for the seams: a small interface, the stdlib `io/fs.FS`
abstraction, and ctx-first command exec.
