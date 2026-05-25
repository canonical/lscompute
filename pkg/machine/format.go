package machine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FmtPretty converts any interface to JSON with indentation, for use in logging where better readability is required. Errors are ignored.
func FmtPretty(v interface{}) string {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		// Ignore error
	}
	return string(jsonData)
}

// FmtBytes converts bytes to a printable string with unit
func FmtBytes(bytes uint64) string {
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

func StringToBytes(sizeString string) (uint64, error) {
	var sizeBytes uint64
	var scaling uint64 = 1
	var err error

	if strings.HasSuffix(sizeString, "G") {
		sizeString = strings.TrimSuffix(sizeString, "G")
		scaling = 1024 * 1024 * 1024
	} else if strings.HasSuffix(sizeString, "M") {
		sizeString = strings.TrimSuffix(sizeString, "M")
		scaling = 1024 * 1024
	}

	sizeBytes, err = strconv.ParseUint(sizeString, 10, 64)
	if err != nil {
		return 0, err
	}
	sizeBytes = sizeBytes * scaling

	return sizeBytes, nil
}
