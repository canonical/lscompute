package disk

import (
	"encoding/json"
	"fmt"
	"testing"
)

var testDirs = []string{
	"/",
	"/var/lib/snapd/snaps",
}

func TestDirStats(t *testing.T) {
	for _, dir := range testDirs {
		t.Run(dir, func(t *testing.T) {
			diskStats, err := statFs(dir)
			if err != nil {
				t.Fatal(err)
			}

			t.Log("Total:", fmtBytes(diskStats.Total))
			t.Log("Avail:", fmtBytes(diskStats.Avail))
		})
	}
}

func TestDirStatsNonExistentDir(t *testing.T) {
	_, err := statFs("/path/that/does/not/exist")
	if err == nil {
		t.Fatal("Non existent dir should return error")
	}
}

func TestInfo(t *testing.T) {
	diskInfo, err := Info()
	if err != nil {
		t.Fatal(err)
	}

	jsonData, err := json.MarshalIndent(diskInfo, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(jsonData))
}

// FmtBytes converts bytes to a printable string with unit
func fmtBytes(bytes uint64) string {
	if bytes > 1024*1024*1024*1024 {
		return fmt.Sprintf("%.1fTiB", float64(bytes)/1024/1024/1024/1024)
	} else if bytes > 1024*1024*1024 {
		return fmt.Sprintf("%.1fGiB", float64(bytes)/1024/1024/1024)
	} else if bytes > 1024*1024 {
		return fmt.Sprintf("%.1fMiB", float64(bytes)/1024/1024)
	} else if bytes > 1024 {
		return fmt.Sprintf("%.1fKiB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%d", bytes)
}
