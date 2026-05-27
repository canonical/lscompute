# Machine test data
Each subdirectory represents a single physical machine.  All data lives under a
`machine-root/` sub-directory that mirrors the real Linux filesystem layout —
the same paths the production code reads, just rooted here instead of at `/`.
Tests use `host.Fake(machineRoot)` to drive the full pipeline against this data
without touching the real host.
---
## Directory layout
```
test_data/machines/<machine-name>/
    machine-root/
        proc/cpuinfo                            ← /proc/cpuinfo
        proc/meminfo                            ← /proc/meminfo
        proc/sys/kernel/arch                    ← /proc/sys/kernel/arch
        sys/bus/pci/devices/<slot>/vendor       ← /sys/bus/pci/devices/…
        sys/bus/pci/devices/<slot>/device
        sys/bus/pci/devices/<slot>/class
        sys/bus/pci/devices/<slot>/subsystem_vendor
        sys/bus/pci/devices/<slot>/subsystem_device
        sys/class/kfd/kfd/topology/nodes/       ← AMD GPUs only
        sys/class/drm/renderD<n>                ← AMD GPUs only (relative symlink)
        usr/share/misc/pci.ids                  ← curated mini database (see below)
        run/disk-stats.json                     ← statfs fixture (see below)
        run/clinfo.json                         ← Intel GPU only; clinfo --json output
        run/nvidia-smi/<slot>/memory.total      ← NVIDIA GPU only
        run/nvidia-smi/<slot>/compute_cap       ← NVIDIA GPU only
    lscompute.json                          ← golden file (curated machines only)
```
Every machine directory **must** have a `lscompute.json` golden file.
The full-pipeline test (`TestGetFromMachineDirs`) asserts that `machine.Get()`
produces output that exactly matches the golden file for every machine.
A missing or empty golden file will cause the sub-test to pass trivially but
provide no regression protection — treat that state as a temporary gap to be
closed before merging.
---
## Capturing data from a real machine
Run all commands below on the target machine, then commit the results.  Replace
`MACHINE` with the directory name you have chosen (e.g. `my-laptop`).
```bash
ROOT=test_data/machines/MACHINE/machine-root
mkdir -p "$ROOT"
```
### proc
```bash
mkdir -p "$ROOT/proc/sys/kernel"
cp /proc/cpuinfo "$ROOT/proc/cpuinfo"
cp /proc/meminfo "$ROOT/proc/meminfo"
cp /proc/sys/kernel/arch "$ROOT/proc/sys/kernel/arch"
```
### PCI sysfs
```bash
mkdir -p "$ROOT/sys/bus/pci/devices"
for slot in /sys/bus/pci/devices/*/; do
    name=$(basename "$slot")
    dest="$ROOT/sys/bus/pci/devices/$name"
    mkdir -p "$dest"
    for attr in vendor device class subsystem_vendor subsystem_device; do
        [ -f "$slot/$attr" ] && cp "$slot/$attr" "$dest/$attr"
    done
done
```
### Disk stats
`run/disk-stats.json` is a JSON map of absolute path → `{"total": <bytes>,
"avail": <bytes>}` for each directory that `disk.Info()` watches.  Currently
only `/var/lib/snapd/snaps` is watched.
```bash
mkdir -p "$ROOT/run"
python3 -c "
import os, json
def statfs(p):
    s = os.statvfs(p)
    return {'total': s.f_blocks * s.f_frsize, 'avail': s.f_bavail * s.f_frsize}
data = {'/var/lib/snapd/snaps': statfs('/var/lib/snapd/snaps')}
print(json.dumps(data, indent=2))
" > "$ROOT/run/disk-stats.json"
```
### Intel GPU (optional — only if the machine has an Intel GPU)
```bash
clinfo --json | python3 -m json.tool > "$ROOT/run/clinfo.json"
```
If `clinfo` is not installed (`sudo apt install clinfo`) the pipeline warns
and skips VRAM lookup; no `clinfo.json` is needed in that case.
### NVIDIA GPU (optional — only if the machine has an NVIDIA GPU)
Replace `SLOT` with the PCI slot of the NVIDIA GPU (e.g. `0000:01:00.0`):
```bash
SLOT=0000:01:00.0
mkdir -p "$ROOT/run/nvidia-smi/$SLOT"
nvidia-smi --id=$SLOT --query-gpu=memory.total --format=csv,noheader \
    > "$ROOT/run/nvidia-smi/$SLOT/memory.total"
nvidia-smi --id=$SLOT --query-gpu=compute_cap --format=csv,noheader \
    > "$ROOT/run/nvidia-smi/$SLOT/compute_cap"
```
### AMD GPU (optional — only if the machine has an AMD GPU)
```bash
mkdir -p "$ROOT/sys/class/kfd/kfd/topology/nodes"
cp -ra /sys/class/kfd/kfd/topology/nodes/. \
       "$ROOT/sys/class/kfd/kfd/topology/nodes/"
mkdir -p "$ROOT/sys/class/drm"
for link in /sys/class/drm/renderD*; do
    cp -a "$link" "$ROOT/sys/class/drm/"
done
```
Symlinks inside `machine-root/` **must be relative** — absolute symlinks
escape the fake root and reach the real filesystem.
---
## pci.ids — curated mini database
Without a `pci.ids` file the pipeline warns about unresolved friendly names.
This is harmless for non-golden machines.  For **golden machines** a curated
`machine-root/usr/share/misc/pci.ids` is required so that the golden file
includes friendly names and the assertion does not depend on the database
installed on the developer's machine.
Author it by hand: extract only the vendor/device lines that the machine
actually references.  The following script prints the relevant lines from the
system database (run on any machine that has `pci-ids` or `hwdata` installed):
```bash
python3 << 'EOF'
import os
SYSFS = "test_data/machines/MACHINE/machine-root/sys/bus/pci/devices"
all_devices, all_subs = {}, {}
for slot in os.listdir(SYSFS):
    d = os.path.join(SYSFS, slot)
    def rd(n):
        try: return open(os.path.join(d,n)).read().strip().replace('0x','').lstrip('0') or '0'
        except: return None
    v,dev,sv,sd = rd('vendor'),rd('device'),rd('subsystem_vendor'),rd('subsystem_device')
    if v and dev:
        v,dev = v.lower(),dev.lower()
        all_devices.setdefault(v,set()).add(dev)
        if sv and sd:
            sv,sd = sv.lower(),sd.lower()
            all_subs.setdefault((v,dev),set()).add((sv,sd))
wanted_vendors = set(all_devices) | {sv for subs in all_subs.values() for sv,_ in subs}
cur_v,in_v,cur_d,in_d = None,False,None,False
with open('/usr/share/misc/pci.ids', errors='replace') as f:
    for line in f:
        line = line.rstrip('\n')
        if not line or line.startswith('#'): continue
        if line.startswith('C '): break
        if line.startswith('\t\t'):
            if in_d:
                ids = line.split('  ',1)[0].split()
                if len(ids)==2:
                    sv2,sd2 = ids[0].lower(),ids[1].lower()
                    if (sv2,sd2) in all_subs.get((cur_v,cur_d),set()):
                        print(line)
        elif line.startswith('\t'):
            did = line.split('  ',1)[0].strip().lower()
            if in_v and did in all_devices.get(cur_v,set()):
                cur_d,in_d = did,True; print(line)
            else: in_d = False
        else:
            in_d = False
            vid = line.split('  ',1)[0].strip().lower()
            cur_v = vid
            in_v = vid in wanted_vendors
            if in_v and vid in all_devices: print(line)
            elif in_v: print(line); in_v = False
EOF
```
Device IDs absent from the installed database must be hand-authored; use
`lspci -nn` on the target machine to get the class description and model the
line on neighbouring entries of the same chipset family.
---
## Refreshing an existing machine
Re-run the capture commands above in the machine's `machine-root/` directory.
(The PCI sysfs step removes and rebuilds the device tree automatically.)
Then regenerate the golden file and review the diff before committing:
```bash
go test ./pkg/machine -run TestGetFromMachineDirs/MACHINE -update -v
```
Review the diff before committing to confirm the output changes are expected.

---
## Keeping golden files up to date

After changing the output format or adding new fields, regenerate **all** golden
files at once:
```bash
go test ./pkg/machine -run TestGetFromMachineDirs -update -v
```
To regenerate a single machine's golden file (useful when only one machine's
raw data changed, or for machine names containing `+` which must be escaped):
```bash
go test ./pkg/machine -run TestGetFromMachineDirs/MACHINE -update -v
# For machine names that contain '+', escape the character in the test filter:
go test ./pkg/machine -run 'TestGetFromMachineDirs/i7-2600k\+arc-a580' -update -v
```
Always review the generated diff before committing — verify friendly names,
VRAM, compute capability, and any other fields that depend on external data.

---
## Adding a new machine
1. Capture all data as described above.
2. Author `machine-root/usr/share/misc/pci.ids` (see section below).
3. Generate the golden file:
   ```bash
   touch test_data/machines/MACHINE/lscompute.json
   go test ./pkg/machine -run TestGetFromMachineDirs/MACHINE -update -v
   ```
4. Review the generated `lscompute.json` — verify friendly names, VRAM,
   compute capability, and other additional properties look correct.
5. Commit everything: `machine-root/`, `usr/share/misc/pci.ids`,
   and `lscompute.json`.
