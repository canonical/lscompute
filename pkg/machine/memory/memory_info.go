package memory

import (
	"fmt"
	"io/fs"

	"github.com/canonical/lscompute/pkg/machine/host"
	"github.com/canonical/lscompute/pkg/machine/types"
)

func Info(h host.Host) (types.MemoryInfo, error) {
	data, err := fs.ReadFile(h.FS(), "proc/meminfo")
	if err != nil {
		return types.MemoryInfo{}, fmt.Errorf("reading proc/meminfo: %v", err)
	}
	return parseProcMemInfo(string(data))
}
