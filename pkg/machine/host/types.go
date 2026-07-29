package host

type dirStats struct {
	// Mountpoint is nil when the backing mountpoint could not be determined.
	Mountpoint *string
	Total      uint64
	Avail      uint64
}
