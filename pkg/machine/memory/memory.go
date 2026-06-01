package memory

import (
	"fmt"
	"io/fs"

	"github.com/canonical/lscompute/pkg/machine/host"
)

func Info(h host.Host) (MemoryInfo, error) {
	data, err := fs.ReadFile(h.FS(), "proc/meminfo")
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("reading proc/meminfo: %w", err)
	}
	return parseProcMemInfo(string(data))
}
