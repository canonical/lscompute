package disk

import (
	"fmt"
	"strings"

	"github.com/canonical/lscompute/pkg/machine/constants"
	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

// directories lists the absolute paths whose disk usage we report. The map
// keys in the result preserve the leading slash for display; we strip it
// internally to satisfy the host.Host io/fs path convention.
var directories = []string{
	constants.SnapStoragePath,
}

// Info returns the total size and available size for configured directories,
// using the host's StatFs implementation.
func Info(h host.Host) (map[string]types.DirStats, error) {
	info := make(map[string]types.DirStats, len(directories))
	for _, dir := range directories {
		dirInfo, err := h.StatFs(strings.TrimPrefix(dir, "/"))
		if err != nil {
			return nil, fmt.Errorf("getting directory info for %s: %w", dir, err)
		}
		info[dir] = dirInfo
	}
	return info, nil
}
