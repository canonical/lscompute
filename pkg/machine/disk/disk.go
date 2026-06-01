package disk

import (
	"fmt"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/host"
)

// directories lists the absolute paths whose disk usage we report. The map
// keys in the result preserve the leading slash for display; we strip it
// internally to satisfy the host.Host io/fs path convention.
var directories = []string{
	snapStoragePath,
}

// Info returns the total size and available size for configured directories,
// using the host's StatFs implementation.
func Info(h host.Host) (map[string]DirInfo, error) {
	info := make(map[string]DirInfo, len(directories))
	for _, dir := range directories {
		hostDirInfo, err := h.StatFs(strings.TrimPrefix(dir, "/"))
		if err != nil {
			return nil, fmt.Errorf("getting directory info for %s: %w", dir, err)
		}

		info[dir] = DirInfo{
			Total: hostDirInfo.Total,
			Avail: hostDirInfo.Avail,
		}
	}
	return info, nil
}
