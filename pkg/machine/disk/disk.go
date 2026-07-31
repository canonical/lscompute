package disk

import (
	"fmt"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func Info(h host.Host) ([]Disk, error) {
	return infoWithDirs(h, h.GetDirectories())
}

// Info returns the total size and available size for configured directories,
// using the host's StatFs implementation.
func infoWithDirs(h host.Host, dirs []string) ([]Disk, error) {
	disks := []Disk{}
	for _, dir := range dirs {
		stat, err := h.DirStat(dir)
		if err != nil {
			return nil, fmt.Errorf("getting directory info for %s: %w", dir, err)
		}

		// mountPoint is best-effort: when it cannot be determined it is left nil.
		var mountPoint *string
		if mp, err := h.GetMountPoint(dir); err == nil {
			mountPoint = &mp
		}

		disks = append(disks, Disk{
			MountPoint: mountPoint,
			Path:       dir,
			Total:      stat.Total,
			Available:  stat.Available,
		})
	}
	return disks, nil
}
