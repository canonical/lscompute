package disk

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/host"
	"golang.org/x/sys/unix"
)

func Info(h host.Host) ([]Disk, error) {
	return infoWithDirs(h, h.GetDirectories())
}

// Info returns the total size and available size for configured directories,
// using the host's StatFs implementation.
func infoWithDirs(h host.Host, dirs []string) ([]Disk, error) {
	disks := []Disk{}
	for _, dir := range dirs {
		buf, err := h.DirStat(dir)
		if err != nil {
			return nil, fmt.Errorf("getting directory info for %s: %w", dir, err)
		}
		st := unix.Statfs_t{
			Blocks: buf.Blocks * uint64(buf.Bsize),
			Bavail: buf.Bavail * uint64(buf.Bsize),
		}

		// mountpoint is best-effort: when it cannot be determined it is left nil.
		var mountpoint *string
		if mp, err := h.GetMountpoint(dir); err == nil {
			mountpoint = &mp
		}

		disks = append(disks, Disk{
			MountPoint: mountpoint,
			Path:       dir,
			Total:      st.Blocks,
			Available:  st.Bavail,
		})
	}
	return disks, nil
}
