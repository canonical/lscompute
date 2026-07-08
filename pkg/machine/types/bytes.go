package types

import "fmt"

// FormatBytes renders a byte count for human-readable YAML output. Values of at
// least one mebibyte are rendered using IEC binary magnitudes with M, G, or T
// suffixes (mebibytes, gibibytes, tebibytes); smaller values are returned as a
// raw byte count.
func FormatBytes(b uint64) any {
	const (
		mib = 1024 * 1024
		gib = 1024 * mib
		tib = 1024 * gib
	)
	switch {
	case b >= tib:
		return fmt.Sprintf("%.1fT", float64(b)/tib)
	case b >= gib:
		return fmt.Sprintf("%.1fG", float64(b)/gib)
	case b >= mib:
		return fmt.Sprintf("%.1fM", float64(b)/mib)
	default:
		return b
	}
}
