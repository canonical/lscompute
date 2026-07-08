package machine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Format is an output serialization format for machine information.
type Format string

const (
	// FormatPlain is a human-readable output with disk and memory capacities
	// rendered using IEC binary suffixes (KiB, MiB, GiB, TiB).
	FormatPlain Format = "plain"
	// FormatJSON is a machine-readable, indented JSON output with kebab-cased keys.
	FormatJSON Format = "json"
)

// Marshal serializes the given machine information using the requested format.
// It returns an error if the format is not recognized or if info is nil.
func Marshal(info *MachineInfo, f Format) ([]byte, error) {
	if info == nil {
		return nil, fmt.Errorf("cannot marshal nil machine info")
	}
	switch f {
	case FormatJSON:
		return json.MarshalIndent(info, "", "  ")
	case FormatPlain:
		return marshalPlain(info)
	default:
		return nil, fmt.Errorf("unknown format %q (choices: plain, json)", f)
	}
}

// marshalPlain renders machine information as human-readable plain text.
func marshalPlain(info *MachineInfo) ([]byte, error) {
	var b strings.Builder

	b.WriteString("CPUs:\n")
	if len(info.Cpus) == 0 {
		b.WriteString("  (none)\n")
	} else {
		block, err := yaml.Marshal(info.Cpus)
		if err != nil {
			return nil, fmt.Errorf("rendering cpus: %w", err)
		}
		b.WriteString(indentLines(string(block), 2))
	}

	b.WriteString("Memory:\n")
	fmt.Fprintf(&b, "  total-ram: %s\n", fmtBytes(info.Memory.TotalRam))
	fmt.Fprintf(&b, "  total-swap: %s\n", fmtBytes(info.Memory.TotalSwap))

	b.WriteString("Disk:\n")
	if len(info.Disk) == 0 {
		b.WriteString("  (none)\n")
	} else {
		paths := make([]string, 0, len(info.Disk))
		for path := range info.Disk {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		for _, path := range paths {
			dir := info.Disk[path]
			fmt.Fprintf(&b, "  %s:\n", path)
			fmt.Fprintf(&b, "    total: %s\n", fmtBytes(dir.Total))
			fmt.Fprintf(&b, "    avail: %s\n", fmtBytes(dir.Avail))
		}
	}

	b.WriteString("Devices:\n")
	if len(info.Devices) == 0 {
		b.WriteString("  (none)\n")
	} else {
		block, err := yaml.Marshal(info.Devices)
		if err != nil {
			return nil, fmt.Errorf("rendering devices: %w", err)
		}
		b.WriteString(indentLines(string(block), 2))
	}

	return []byte(b.String()), nil
}

// fmtBytes converts bytes to a printable string using IEC binary prefixes
// (KiB, MiB, GiB, TiB). Values below 1KiB are rendered as a raw byte count.
func fmtBytes(bytes uint64) string {
	if bytes >= 1024*1024*1024*1024 {
		return fmt.Sprintf("%.1fTiB", float64(bytes)/1024/1024/1024/1024)
	} else if bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.1fGiB", float64(bytes)/1024/1024/1024)
	} else if bytes >= 1024*1024 {
		return fmt.Sprintf("%.1fMiB", float64(bytes)/1024/1024)
	} else if bytes >= 1024 {
		return fmt.Sprintf("%.1fKiB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%d", bytes)
}

// indentLines prefixes each non-empty line of s with the given number of spaces.
func indentLines(s string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	var b strings.Builder
	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(prefix)
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
