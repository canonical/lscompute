package disk

import (
	"fmt"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// directories lists the absolute paths whose disk usage we report. Each path is
// stripped of its leading slash internally to satisfy the host.Host io/fs path
// convention; the resolved mountpoint is reported for each entry.
var directories = []string{
	snapStoragePath,
}

// Info returns the total size and available size for configured directories,
// using the host's StatFs implementation.
func Info(h host.Host) ([]Disk, error) {
	disks := []Disk{}
	for _, dir := range directories {
		hostDirInfo, err := h.StatFs(strings.TrimPrefix(dir, "/"))
		if err != nil {
			return nil, fmt.Errorf("getting directory info for %s: %w", dir, err)
		}

		disks = append(disks, Disk{
			MountPoint: hostDirInfo.Mountpoint,
			Path:       dir,
			Total:      hostDirInfo.Total,
			Available:  hostDirInfo.Avail,
		})
	}
	return disks, nil
}
