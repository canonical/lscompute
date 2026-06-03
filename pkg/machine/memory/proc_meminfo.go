package memory

import (
	"fmt"
	"strconv"
	"strings"
)

func parseProcMemInfo(memInfoString string) (MemoryInfo, error) {
	var memInfo MemoryInfo
	foundMemTotal := false

	lines := strings.Split(memInfoString, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		fields := strings.SplitN(line, ":", 2)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0]) // remove \t between key and colon
		value := strings.TrimSpace(fields[1])

		switch key {
		case "MemTotal":
			valueBytes, err := procStringToBytes(value)
			if err != nil {
				return memInfo, fmt.Errorf("parsing MemTotal: %w", err)
			}
			memInfo.TotalRam = uint64(valueBytes)
			foundMemTotal = true
		case "SwapTotal":
			valueBytes, err := procStringToBytes(value)
			if err != nil {
				return memInfo, fmt.Errorf("parsing SwapTotal: %w", err)
			}
			memInfo.TotalSwap = uint64(valueBytes)
		}
	}

	if !foundMemTotal {
		return memInfo, fmt.Errorf("required field MemTotal not found")
	}

	return memInfo, nil
}

func procStringToBytes(s string) (int64, error) {
	s = strings.TrimSpace(s)

	if strings.HasSuffix(s, "kB") {
		s = strings.TrimSuffix(s, "kB")
		s = strings.TrimSpace(s)
		kbValue, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parsing kB value: %w", err)
		}
		return kbValue * 1024, nil
	} else {
		bValue, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parsing byte value: %w", err)
		}
		return bValue, nil
	}
}
